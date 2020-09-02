// Package png allows for loading png images and applying
// image flitering effects on them.
package png

import (
	"image/color"
)

// Grayscale applies a grayscale filtering effect to the image
func (img *Image) Grayscale() {

	// Bounds returns defines the dimensions of the image. Always
	// use the bounds Min and Max fields to get out the width
	// and height for the image
	bounds := img.out.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Returns the pixel (i.e., RGBA) value at a (x,y) position
			// Note: These get returned as int32 so based on the math you'll
			// be performing you'll need to do a conversion to float64(..)
			r, g, b, a := img.in.At(x, y).RGBA()

			// Note: The values for r,g,b,a for this assignment will range between [0, 65535].
			// For certain computations (i.e., convolution) the values might fall outside this
			// range so you need to clamp them between those values.
			greyC := clamp(float64(r+g+b) / 3)

			// Note: The values need to be stored back as uint16 (I know weird..but there's valid reasons
			// for this that I won't get into right now).
			img.out.Set(x, y, color.RGBA64{greyC, greyC, greyC, uint16(a)})
		}
	}
}

// Convolute performs a sharpen, edge-detection, or blur effect based on the kernel given.
func (img *Image) Convolute(kernel [][]float64, kernelLength int) {

	// For each pixel in the image do the convolution.
	bounds := img.out.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b, a uint32
			var rConvoluted, gConvoluted, bConvoluted float64 
			_, _, _, a = img.in.At(x, y).RGBA()

			// Loop through the kernel and multiply kernel rgba value by corresponding image rgba value.
			// Pad with zeros for edges.
			for kernelY := 0; kernelY < kernelLength; kernelY++ {
				for kernelX := 0; kernelX < kernelLength; kernelX++ {
					xOfImage := x + kernelX - 1
					yOfImage := y + kernelY - 1
					if xOfImage < 0 || yOfImage < 0 {
						r = 0
						g = 0
						b = 0
					} else {
						r, g, b, _ = img.in.At(xOfImage, yOfImage).RGBA()
					}
					// Accumulate the r, g, b values for a given image pixel.
					rConvoluted += float64(r) * kernel[kernelY][kernelX]
					gConvoluted += float64(g) * kernel[kernelY][kernelX]
					bConvoluted += float64(b) * kernel[kernelY][kernelX]
				}
			}

			// Save new pixel to out image at same image pixel position as the original.
			filteredPixelColor := color.RGBA64{clamp(rConvoluted), clamp(gConvoluted), clamp(bConvoluted), uint16(a)}
			img.out.Set(x, y, filteredPixelColor)
		}
	}
}