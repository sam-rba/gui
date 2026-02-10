package layout

import (
	"image"
	"image/draw"
)

type subimager interface {
	SubImage(image.Rectangle) image.Image
}

// Subimage returns an image representing the portion of the image m visible through r.
// The returned value shares pixels with the original.
// Panics if the concrete image type does not have a SubImage() method.
func subimage(m draw.Image, r image.Rectangle) draw.Image {
	return m.(subimager).SubImage(r).(draw.Image)
}
