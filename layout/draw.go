package layout

import (
	"image"
	"image/color"
	"image/draw"
)

// subimager is implemented by most Image types.
type subimager interface {
	SubImage(image.Rectangle) image.Image
}

// Subimage returns an image representing the portion of the image m visible through r.
// The returned value shares pixels with the original.
// Panics if the concrete image type does not have a SubImage() method.
func subimage(m draw.Image, r image.Rectangle) draw.Image {
	return m.(subimager).SubImage(r).(draw.Image)
}

// drawSubimage translates a draw call onto the given subimage area.
func drawSubImage(f func(draw.Image) image.Rectangle, r image.Rectangle) func(draw.Image) image.Rectangle {
	return func(img draw.Image) image.Rectangle {
		return f(subimage(img, r))
	}
}

// drawBackground returns a draw call that fills the entire image with a color.
func drawBackground(c color.Color) func(draw.Image) image.Rectangle {
	return func(img draw.Image) image.Rectangle {
		draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.ZP, draw.Src)
		return img.Bounds()
	}
}
