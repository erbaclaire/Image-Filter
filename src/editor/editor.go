package main

import (
	"os"
	"strconv"
	"fmt"
	"runtime"
	"bufio"
	"strings"
	"src/imagetask"
)

// printUsage prints the required and optional command line arguments that should / can be specified for running the program.
func printUsage() {
	fmt.Println("Usage: editor [-p=[number of threads]]\n\t-p=[number of threads] = An optional flag to run the editor in its parallel version.\n\tCall and pass the runtime.GOMAXPROCS(...) function the integer specified by [number of threads].")
}

// worker is a goroutine that reads from the image task channel.
// worker grabs an image task from the image task channel, splits it up, appliesfilters to the portions, and then recombines the portions.
// worker then pushes the filtered total image to the results channel for saving.
// worker will continue to grab tasks until there are no my image tasks to consumer and then it will exit.
func worker(threads int, doneWorker chan<- bool, imageTasks <-chan *imagetask.ImageTask, imageResults chan<- *imagetask.ImageTask) {

	// Initial placing image chunks in to a channel for filter to consume.
	chunkStreamGenerator := func(done <- chan interface{}, imageChunks []*imagetask.ImageTask) chan *imagetask.ImageTask {
		chunkStream := make(chan *imagetask.ImageTask)
		go func() {
			defer close(chunkStream)
			for _, chunk := range imageChunks {
				select {
				case <-done:
					return
				case chunkStream <- chunk:
				}
			}
		}()
		return chunkStream
	}

	// Filter applies a filter in a pipeline fashion. 
	// A goroutine is spawned for each chunk that needs to be filtered (which is numOfThreads chunks for each filter effect)
	filter := func(threads int, effect string, effectNum int, done <- chan interface{}, chunkStream chan *imagetask.ImageTask) chan *imagetask.ImageTask {
		filterStream := make(chan *imagetask.ImageTask, threads) // Only numOfThreads image chunks should be in the local filter channel.
		donefilterChunk := make(chan bool)
		for chunk := range chunkStream { // For each image chunk ...
			if effectNum > 0 {
				chunk.Img.UpdateInImg() // Replace inImg with outImg if not the first effect to compund effects.
			}	
			go func(chunk *imagetask.ImageTask) { // Spawn a goroutine for each chunk, which is equal to the numOfThreads. Each goroutine works on a portion of the image.
				select {
				case <-done:
					donefilterChunk <- true
					return
				case filterStream <- chunk:
					if effect != "G" {
						chunk.Img.ApplyConvolution(effect) // Can wait to apply effect until after chunk is in the channel because has to wait for all goroutines to finish before it can move on to the next filter for a given image.
					} else {
						chunk.Img.Grayscale()
					}
					donefilterChunk <- true // Indicate that the filtering is done for the given chunk.
				}
			}(chunk)
		}
		for i := 0; i < threads; i ++ { // Wait for all portions to be put through one filter because of image dependencies with convolution.
			<-donefilterChunk
		}
		return filterStream
	}

	// While there are more image tasks to grab ...
	for true {
		// Grab image task from image task channel.	
		imgTask, more := <-imageTasks

		// If you get an image task, split up the image in to even chunks by y-pixels.
		if more {
			imageChunks := imgTask.SplitImage(threads)
			
			// Iterate through filters on image chunks.
			// Will spawn a goroutine for each chunk in each filter (n goroutines per filter)
			done := make(chan interface{})
			defer close(done)
			chunkStream := chunkStreamGenerator(done, imageChunks)
			for i := 0; i < len(imgTask.Effects); i++ {
				effect := imgTask.Effects[i]
				chunkStream = filter(threads, effect, i, done, chunkStream)
				close(chunkStream)
			}

			// Put the image back together.
			reconstructedImage, _ := imgTask.Img.NewImage()
			for imgChunk := range chunkStream {
				reconstructedImage.ReAddChunk(imgChunk.Img, imgChunk.YPixelStart, imgChunk.ChunkPart)
			}
			imgTask.Img = reconstructedImage
			imageResults <- imgTask // Send image to results channel to be saved.

		} else { // Otherwise, if there are no more image tasks, then goroutine worker exits.
			doneWorker <- true // Indicate that the worker is done.
			return
		}
	}
}

// main determines whether the sequential or parallel version of the program should be run based on the command line argument.
// main spawns a goroutine to read in standard input and puts image tasks if a channel for consumption for workers.
// main spawns n worker goroutines to consume image tasks.
// main spawns a goroutine to read image results after the filtering process has completed and saves those images.
func main() {

	// If invalid arguments then print usage statement.
	if len(os.Args) < 1 || len(os.Args) > 2 {

		printUsage()

	} else if len(os.Args) == 2 { // Run parallel version.
		
		threads, _ := strconv.Atoi(strings.Split(os.Args[1],"=")[1])
		runtime.GOMAXPROCS(threads)
		imageTasks := make(chan *imagetask.ImageTask)
		imageResults := make(chan *imagetask.ImageTask)
		doneWorker := make(chan bool)
		doneSave := make(chan bool)

		// Read from standard input (initial generator).
		// Image tasks go to imageTasks channel which is consumed by multiple workers (fan out).
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				imageTaskText := scanner.Text()
				imgTask := imagetask.CreateImageTask(imageTaskText)
				imageTasks <- imgTask
			}
			close(imageTasks)
		}()

		// Spawn worker goroutines to split up images, filter the portions, and recombine images.
		for i := 0; i < threads; i++ {
			go worker(threads, doneWorker, imageTasks, imageResults)
		}

		// Save results.
		go func() {
			for true { // Do while there are more images to save.
				imgTask, more := <- imageResults // Reads from the image results channel.
				if more {
					imgTask.SaveImageTaskOut()
				} else {
					doneSave <- true 
					return
				}
			}
		}()

		// Wait for all workers to return.
		for i := 0; i < threads; i++ {
			<-doneWorker
		}

		// Wait for all images to be saved.
		close(imageResults)
		<- doneSave

	} else { // Run sequential version if '-p' flag not given.

		// Read in image tasks from standard input.
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			imageTaskText := scanner.Text()
			imgTask := imagetask.CreateImageTask(imageTaskText)

			// Apply the specified effects to the given image.
			for i, effect := range imgTask.Effects {
				if i > 0 {
					imgTask.Img.UpdateInImg()
				}
				if effect != "G" {
					imgTask.Img.ApplyConvolution(effect)
				} else {
					imgTask.Img.Grayscale()
				}
			}

			// Save filtered image.
			imgTask.SaveImageTaskOut()

		}

	}

}