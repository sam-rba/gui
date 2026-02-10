package layout

import "image"

// ResizeAll resizes a layout to consume all of its parent's area.
func ResizeAll(r image.Rectangle) image.Rectangle { return r }

// ResizeQuad1 resizes a layout to consume the top-right quadrant of its parent.
func ResizeQuad1(r image.Rectangle) image.Rectangle {
	mid := midpoint(r)
	return image.Rect(mid.X, r.Min.Y, r.Max.X, mid.Y)
}

// ResizeQuad2 resizes a layout to consume the top-left quadrant of its parent.
func ResizeQuad2(r image.Rectangle) image.Rectangle {
	mid := midpoint(r)
	return image.Rectangle{r.Min, mid}
}

// ResizeQuad3 resizes a layout to consume the bottom-left quadrant of its parent.
func ResizeQuad3(r image.Rectangle) image.Rectangle {
	mid := midpoint(r)
	return image.Rect(r.Min.X, mid.Y, mid.X, r.Max.Y)
}

// ResizeQuad4 resizes a layout to consume the bottom-right quadrant of its parent.
func ResizeQuad4(r image.Rectangle) image.Rectangle {
	mid := midpoint(r)
	return image.Rectangle{mid, r.Max}
}

// midpoint returns the point in the middle of a rectangle.
func midpoint(r image.Rectangle) image.Point {
	return r.Min.Add(r.Max).Div(2)
}
