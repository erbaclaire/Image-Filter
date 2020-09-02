### Many algorithms in image processing benefit from parallelization. I created an image processing system that reads in a series of images and applies certain effects to them using image convolution. If you are unfamiliar with image convolution then you should read over the following sources before continuing:
* http://www.songho.ca/dsp/convolution/convolution2d_example.html
* https://en.wikipedia.org/wiki/Kernel_(image_processing)

## Image Processing System
* I created an image editor that applies image effects on a series of images using 2D image convolution. The program reads in from os.Stdin JSON strings, where each string represents an image along with the effects that should be applied to that image. Each string will have the following format:
```{"inPath": string, "outPath": string, "effects": [string]}```
	* The "inPath" pairing represents the file path of the image to read in. Images will always be PNG files.
	* The "outPath" pairing represents the file path to save the image after applying the effects.
 	* The "effects" pairing represents the image effects to apply to the IMAGE IN PATH image. Effects are applied in the order they are listed. If no effects are specified (e.g., []) then the out image is the same as the input image. 
* The program will read in the images, apply the effects associated with an image, and save the images to their specified output file paths. How the program processes this file is described in the Program Specifica- tions section.

## Image Effects
* The sharpen, edge-detection, and blur image effects are required to use image convolution to apply their effects to the input image. Again, read about convolution here:
http://www.songho.ca/dsp/convolution/convolution2d_example.html
	* As stated in the above article, the size of the input and output image are fixed (i.e., they are the same).
	* Thus, results around the border pixels will not be fully accurate because you will need to pad zeros where inputs are not defined. 
* The grayscale effect uses a simple algorithm defined below that does not require convolution.

### Image Effect Effect Description
* S - Performs a sharpen effect with the following kernel: ```[[0 −1 0], [−1 5 −1], [0 −1 0]]```
* E - Performs a edge-detection effect with the following kernel: ```[[−1 −1 −1], [−1 8 −1], [−1 −1 −1]]```
* B Performs a blur effect with the following kernel: ```1/9 [[1 1 1], [1 1 1], [1 1 1]]```
* G - Performs a grayscale effect on the image. This is done by averaging the values of all three color num- bers for a pixel, the red, green and blue, and then replacing them all by that average. So if the three colors were 25, 75 and 250, the average would be 116, and all three numbers would become 116.

## Program Specifications
* There are two versions of the editor program: a sequential version and a parallel version. 
* The program has the following usage statement: ```Usage: editor [-p=[number of threads]]``` where ```-p=[number of threads] = An optional flag to run the editor in its parallel version.```
* Program calls and passes the runtime.GOMAXPROCS(...) function the integer specified by [number of threads].

## Sequential Version
* The sequential version is ran by default when executing the editor program. The user must provide the -p flag to specify that they want to run the program’s parallel version. The sequential program is relatively straightforward. As described above, this version runs through the images specified by the strings coming in from os.Stdin, applies their effects and saves the modified images to their output files.

## Parallel Version
* The parallel version is ran with the -p flag. The parallel implementation is a mixture of using functional decomposition and data decomposition. The implementation is as follows:
1. The integer given to the -p flag is passed to the runtime.GOMAXPROCS(numOfThreads) function. 
2. All synchronization between the goroutines is done using channels. Nothing from the sync package is used.
3. There is a fanin/fanout scheme as follows:
* Image Task Generator: As stated earlier, the program will read in the images to process via os.Stdin. Reading from the os.Stdin is done by a single generator goroutine. The image task generator reads in the JSON strings and does any preparation needed before applying their effects. The output from this goroutine is an ImageTask value. The image task generator writes each ImageTask to a channel and multiple workers will read from it.
* Workers: The workers are the goroutines that are performing the filtering effects on the images. The number of workers is static and is equal to the numOfThreads command line argument. A worker uses a pipeline pattern. Each stage of the pipeline, has a data decomposition component, which does the following:
	* Spawn N number of goroutines, where N = numOfThreads.
	* Each spawned goroutine is given a section of the image to work on.
	* Each spawned goroutine applies the effect for that stage to its assigned section.
* Results Aggregator: The results aggregator gorountine reads from the channel that holds the ImageResults and saves the filtered image to its “outpath” file.
4. If all the images have been processed then the main goroutine can exit the program.

## Testing
* You can find test images and csv files here: https://www.dropbox.com/s/s6sws5w4xcnx94e/proj2_files.zip?dl=0
* To test, place the contents of file_1, file_2, or file_3 directly in the /src/editor directory and run the usage statement with file_1.txt, file_2.txt, or file_3.txt as the input. For instance, ```go run editor.go -p=6 < file_1.txt```.
* See report.pdf for an analysis of the efficiencies gained from the parallel version.
