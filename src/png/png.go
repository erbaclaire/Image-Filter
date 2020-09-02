// Package png allows for loading png images and applying
// image flitering effects on them
package png

import (
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
)

// The Image represents a structure for working with PNG images.
type Image struct {
	in       image.Image
	out      *image.RGBA64
}

// SubImager interface to be able to grab a subsection of the image pixels.
type SubImager interface {
    image.Image
    SubImage(r image.Rectangle) image.Image
}

//
// Public functions
//

// Load returns a Image that was loaded based on the filePath parameter
func Load(filePath string) (*Image, error) {

	inReader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer inReader.Close()
	inImg, err := png.Decode(inReader)
	if err != nil {
		return nil, err
	}

	inBounds := inImg.Bounds()
	outImg := image.NewRGBA64(inBounds)

	return &Image{inImg, outImg}, nil
}

// Save saves the image to the given file
func (img *Image) Save(filePath string, noChange bool) error {

	outWriter, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outWriter.Close()

	// If no effects save the in-image otherwise save the out-image.
	if noChange {
		err = png.Encode(outWriter, img.in)
		if err != nil {
			return err
		}
	} else {
		err = png.Encode(outWriter, img.out)
		if err != nil {
			return err
		}
	}

	return nil
}

// clamp will clamp the comp parameter to zero if it is less than zero or to 65535 if the comp parameter
// is greater than 65535.
func clamp(comp float64) uint16 {
	return uint16(math.Min(65535, math.Max(0, comp)))
}

// getSKernel returns the kernel matrix used for the sharpen effect convolution.
func getSKernel() [][]float64 {
	kernel := [][]float64{
		{0, -1, 0},
		{-1, 5, -1},
		{0, -1, 0},
	}
	return kernel 
}

// getEKernel returns the kernel matrix used for the edge-detection effect convolution.
func getEKernel() [][]float64 {
	kernel := [][]float64{
		{-1, -1, -1},
		{-1, 8, -1},
		{-1, -1, -1},
	}
	return kernel 
}

// getBKernel returns the kernel matrix used for the blur effect convolution.
func getBKernel() [][]float64 {
	var row []float64
	var kernel [][]float64
	oneNinth := float64(1)/float64(9)
	for i := 0; i < 3; i++ {
		row = append(row, oneNinth)
	}
	for i := 0; i < 3; i++ {
		kernel = append(kernel, row)
	}
	return kernel 
}

// ApplyConvolution applies the sharpen, edge-detection, or blur on the image by applying the correct kernel filter.
func (img *Image) ApplyConvolution(effect string) {
	var kernel [][]float64
	if effect == "S" {
		kernel = getSKernel()
	} else if effect == "E" {
		kernel = getEKernel()
	} else if effect == "B" {
		kernel = getBKernel()
	}
	img.Convolute(kernel, len(kernel))	
}

// UpdateInImg updates the img in value so that anoth effect can be compunded on top of the previous one.
func (img *Image) UpdateInImg() {
	img.in = img.out
	inBounds := img.in.Bounds()
	img.out = image.NewRGBA64(inBounds)
}

// GetYPixels returns the number of y-axis pixels for the purposes of splitting up the image.
func (img *Image) GetYPixels() int {
	return img.in.Bounds().Max.Y
}

// GrabChunk grabs a chunk of the original image given a min and max Y pixel number.
// GrabChunk essentially grabs a subset of pixel rows from the image.
func (img *Image) GrabChunk(yMin int, yMax int) (*Image, error) {
	xMax := img.in.Bounds().Max.X
	var subImg image.Image
	if p, ok := img.in.(SubImager); ok {
		subImg = p.SubImage(image.Rect(0, yMin, xMax, yMax))
	}
	inBounds := subImg.Bounds()
	outImg := image.NewRGBA64(inBounds)
	return &Image{subImg, outImg}, nil
}

// NewImage creates a new blank canvas to "put together" all of the individual image chunks.
func (img *Image) NewImage() (*Image, error) {
	xMax := img.in.Bounds().Max.X
	yMax := img.in.Bounds().Max.Y
	newImage := image.NewRGBA64(image.Rect(0, 0, xMax, yMax))
	return &Image{img.in, newImage}, nil
}

// ReAddChunk puts the image chunks back together after all of the filtering.
func (img *Image) ReAddChunk(img2 *Image, yMin int, chunkPart string) {
	// Fix edge cases from convolution.
	bounds := img2.out.Bounds()
	if chunkPart == "first" {
		bounds.Max.Y -= 4
	} else if chunkPart == "middle" {
		bounds.Min.Y += 4
		yMin += 4
		bounds.Max.Y -= 4
	} else if chunkPart == "end" {
		bounds.Min.Y += 4
		yMin += 4
	} 
	// Get cropped version of img2 so that edge cases of portions do not overlap.
	subImg := img2.out.SubImage(image.Rect(0, yMin, bounds.Max.X, bounds.Max.Y))
	draw.Draw(img.out, bounds, subImg, image.Point{0, yMin}, draw.Src)
}
