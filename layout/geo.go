package layout

import (
	"fmt"
	"image"
)

// ResizeFunc is a function that computes the area of a layout when it receives a Resize event.
// It takes the Resize's Rectangle and returns the area that the layout should cover.
type ResizeFunc func(image.Rectangle) image.Rectangle

// Full resizes a layout to consume all of its parent's area.
func Full() ResizeFunc {
	return func(r image.Rectangle) image.Rectangle { return r }
}

// Quad1 resizes a layout to consume the top-right quadrant of its parent.
func Quad1() ResizeFunc {
	return func(r image.Rectangle) image.Rectangle {
		mid := midpoint(r)
		return image.Rect(mid.X, r.Min.Y, r.Max.X, mid.Y)
	}
}

// Quad2 resizes a layout to consume the top-left quadrant of its parent.
func Quad2() ResizeFunc {
	return func(r image.Rectangle) image.Rectangle {
		mid := midpoint(r)
		return image.Rectangle{r.Min, mid}
	}
}

// Quad3 resizes a layout to consume the bottom-left quadrant of its parent.
func Quad3() ResizeFunc {
	return func(r image.Rectangle) image.Rectangle {
		mid := midpoint(r)
		return image.Rect(r.Min.X, mid.Y, mid.X, r.Max.Y)
	}
}

// Quad4 resizes a layout to consume the bottom-right quadrant of its parent.
func Quad4() ResizeFunc {
	return func(r image.Rectangle) image.Rectangle {
		mid := midpoint(r)
		return image.Rectangle{mid, r.Max}
	}
}

// Inset resizes a layout to be inset some number of pixels from the edges of its parent.
func Inset(top, right, bottom, left int) ResizeFunc {
	if top < 0 {
		panic(fmt.Sprintf("inset %d < 0", top))
	} else if right < 0 {
		panic(fmt.Sprintf("inset %d < 0", right))
	} else if bottom < 0 {
		panic(fmt.Sprintf("inset %d < 0", bottom))
	} else if left < 0 {
		panic(fmt.Sprintf("inset %d < 0", left))
	}
	return func(r image.Rectangle) image.Rectangle {
		if r.Dx() < left+right {
			r.Min.X = (r.Min.X + r.Max.X) / 2
			r.Max.X = r.Min.X
		} else {
			r.Min.X += left
			r.Max.X -= right
		}
		if r.Dy() < top+bottom {
			r.Min.Y = (r.Min.Y + r.Max.Y) / 2
			r.Max.Y = r.Min.Y
		} else {
			r.Min.Y += top
			r.Max.Y -= bottom
		}
		return r
	}
}

// InsetAll resizes as layout to be inset px pixels on all sides from the edges of its parent.
func InsetAll(px int) ResizeFunc { return Inset(px, px, px, px) }

// InsetTop resizes a layout to be inset px pixels from the top edge of its parent.
func InsetTop(px int) ResizeFunc { return Inset(px, 0, 0, 0) }

// InsetRight resizes a layout to be inset px pixels from the top edge of its parent.
func InsetRight(px int) ResizeFunc { return Inset(0, px, 0, 0) }

// InsetBottom resizes a layout to be inset px pixels from the top edge of its parent.
func InsetBottom(px int) ResizeFunc { return Inset(0, 0, px, 0) }

// InsetLeft resizes a layout to be inset px pixels from the top edge of its parent.
func InsetLeft(px int) ResizeFunc { return Inset(0, 0, 0, px) }

// midpoint returns the point in the middle of a rectangle.
func midpoint(r image.Rectangle) image.Point {
	return r.Min.Add(r.Max).Div(2)
}
