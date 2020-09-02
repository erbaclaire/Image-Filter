package imagetask

import (
	"src/png"
	"math"
	"encoding/json"
)

// ImageTask represents the structure of incoming JSON from standard input
// alongside other fields to help with convolution in the parallel implementation.
type ImageTask struct {
	InPath   	   string         `json:"inPath"`
	OutPath		   string         `json:"outPath"`  
	Effects        []string       `json:"effects"`
	Img            *png.Image     `json:"img"`  // the entire image.
	YPixelStart    int 	          `json:"ypixelstart"` // original y-pixel from image when chunk started.
	YPixelEnd      int 	          `json:"ypixelend"` // original y-pixel from image when chunk ended.
	ChunkPart	   string     	  `json:"chunkpart"` // string to identify if the portion of the image is the first chunk, last chunk, or one of the middle for putting back together.
}

// readInputJSON reads in image data for a certain image filter task and returns the image task with the image pixel data attached.
func CreateImageTask(imageTask string) *ImageTask {
	imageTaskJSONBytes := []byte(imageTask)
	var imgTask ImageTask
	err := json.Unmarshal(imageTaskJSONBytes, &imgTask)
	if err != nil {
		panic(err)
	}
	pngImg, err := png.Load(imgTask.InPath)
	if err != nil {
		panic(err)
	}	
	imgTask.Img = pngImg
	return &imgTask
}

// savePic saves the final, filtered picture to the outPath.
func (imgTask *ImageTask) SaveImageTaskOut() {
	var saveOld bool
	if len(imgTask.Effects) == 0 {
		saveOld = true
	} else {
		saveOld = false
	}
	err := imgTask.Img.Save(imgTask.OutPath, saveOld)
	if err != nil {
		panic(err)
	}
}

// splitImage splits the image in to equal chunks for filtering.
func (imgTask *ImageTask) SplitImage(threads int) []*ImageTask {
	var imageChunks []*ImageTask
	yPixels := imgTask.Img.GetYPixels()
	pixelChunk := int(math.Ceil(float64(yPixels)/float64(threads)))
	yChunkMin := 0
	yChunkMax := 0
	var portion string
	for i := 0; i < threads; i++ {
		// Grab next chunk of image.
		yChunkMin = i*pixelChunk
		if i == threads-1 { // Grab the rest of the image in case it doesn't divide into exactly even chunks for last thread.
			yChunkMax = yPixels
		} else {
			yChunkMax = yChunkMin + pixelChunk
		}
		// Deal with edges for convolution - give padding so that the edges of chunks are not off.
		if threads > 1 {
			if i == 0 {
				portion = "first"
				yChunkMax += 4
			} else if i > 0 && i < threads-1 {
				portion = "middle"
				yChunkMin -= 4
				yChunkMax += 4
			} else {
				portion = "end"
				yChunkMin -= 4
			}
		} else {
			portion = "na"
		}
		// Grab the chunk.
		chunk, _ := imgTask.Img.GrabChunk(yChunkMin, yChunkMax)
		imageChunks = append(imageChunks, &ImageTask{imgTask.InPath, imgTask.OutPath, imgTask.Effects, chunk, yChunkMin, yChunkMax, portion})
	}
	return imageChunks
}