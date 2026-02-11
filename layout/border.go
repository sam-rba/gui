package layout

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
)

// Border is a layout that draws a border around itself. It can have margin outside
// the border, and padding inside the border. It can also be used just for spacing,
// with only margin/padding and no border.
type Border struct {
	events <-chan gui.Event
	draw   chan<- func(draw.Image) image.Rectangle
}

// NewBorder creates a Border layout with optional margin, border, and padding thicknesses.
func NewBorder(env gui.Env, opts ...BorderOption) gui.Env {
	o := evalBorderOptions(opts...)

	events := make(chan gui.Event)                       // to child
	draws := make(chan func(draw.Image) image.Rectangle) // from child

	go func(events chan<- gui.Event, draws <-chan func(draw.Image) image.Rectangle) {
		defer close(env.Draw())
		defer close(events)

		redraw := func(area image.Rectangle) {
			env.Draw() <- drawBorder(o)
			fmt.Println("area, content:", area, borderContentArea(area, o))
			events <- gui.Resize{borderContentArea(area, o)} // translate and forward to child
		}

		// Forward first resize event to child and draw border
		event := (<-env.Events()).(gui.Resize) // first event guaranteed to be Resize
		area := event.Rectangle
		go redraw(area)

		for {
			select {
			case event := <-env.Events(): // event from parent
				switch event := event.(type) {
				case gui.Resize:
					area = event.Rectangle
					go redraw(area)
				default:
					go func() { events <- event }() // forward to child
				}
			case f, ok := <-draws: // draw call from child
				if !ok {
					return
				}
				env.Draw() <- drawSubImage(f, borderContentArea(area, o))
			}
		}
	}(events, draws)

	return Border{events, draws}
}

// BorderContentArea returns the drawable area inside the margin,
// border, and padding of a border layout, given the overall area
// of the layout and the options.
func borderContentArea(area image.Rectangle, o borderOptions) image.Rectangle {
	top := o.margin.top + o.border.top + o.padding.top
	right := o.margin.right + o.border.right + o.padding.right
	bottom := o.margin.bottom + o.border.bottom + o.padding.bottom
	left := o.margin.left + o.border.left + o.padding.left
	return Inset(top, right, bottom, left)(area)
}

func drawBorder(o borderOptions) func(draw.Image) image.Rectangle {
	return func(img draw.Image) image.Rectangle {
		r := Inset(o.margin.top, o.margin.right, o.margin.bottom, o.margin.left)(img.Bounds())
		mask := image.NewAlpha(image.Rect(0, 0, r.Dx(), r.Dy()))
		draw.Draw(mask, mask.Bounds(), image.Opaque, image.ZP, draw.Src)                                                                           // mask border
		draw.Draw(mask, Inset(o.border.top, o.border.right, o.border.bottom, o.border.left)(mask.Bounds()), image.Transparent, image.ZP, draw.Src) // unmask inside border
		draw.DrawMask(img, r, &image.Uniform{o.Color}, image.ZP, mask, image.ZP, draw.Over)
		fmt.Println("border:", r)
		return r
	}
}

// BorderOption is a functional option for the NewBorder constructor.
type BorderOption func(o *borderOptions)

type borderOptions struct {
	margin, border, padding sides
	color.Color
}

type sides struct{ top, right, bottom, left int }

// MarginAll option sets the margin size (in pixels) on all sides.
func MarginAll(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("margin %d < 0", px))
	}
	return func(o *borderOptions) { o.margin = sides{px, px, px, px} }
}

// MarginTop option sets the top margin size, in pixels.
func MarginTop(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("margin %d < 0", px))
	}
	return func(o *borderOptions) { o.margin.top = px }
}

// MarginRight option sets the right margin size, in pixels.
func MarginRight(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("margin %d < 0", px))
	}
	return func(o *borderOptions) { o.margin.right = px }
}

// MarginBottom option sets the bottom margin size, in pixels.
func MarginBottom(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("margin %d < 0", px))
	}
	return func(o *borderOptions) { o.margin.bottom = px }
}

// MarginLeft option sets the left margin size, in pixels.
func MarginLeft(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("margin %d < 0", px))
	}
	return func(o *borderOptions) { o.margin.left = px }
}

// BorderColor sets the color of the border.
func BorderColor(c color.Color) BorderOption {
	return func(o *borderOptions) { o.Color = c }
}

// BorderAll option sets the border thickness (in pixels) on all sides.
func BorderAll(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("border thickness %d < 0", px))
	}
	return func(o *borderOptions) {
		o.border = sides{px, px, px, px}
	}
}

// BorderTop sets the top border thickness, in pixels.
func BorderTop(thick int) BorderOption {
	if thick < 0 {
		panic(fmt.Sprintf("border thickness %d < 0", thick))
	}
	return func(o *borderOptions) { o.border.top = thick }
}

// BorderRight sets the right border thickness, in pixels.
func BorderRight(thick int) BorderOption {
	if thick < 0 {
		panic(fmt.Sprintf("border thickness %d < 0", thick))
	}
	return func(o *borderOptions) { o.border.right = thick }
}

// BorderBottom sets the bottom border thickness, in pixels.
func BorderBottom(thick int) BorderOption {
	if thick < 0 {
		panic(fmt.Sprintf("border thickness %d < 0", thick))
	}
	return func(o *borderOptions) { o.border.bottom = thick }
}

// BorderLeft sets the left border thickness, in pixels.
func BorderLeft(thick int) BorderOption {
	if thick < 0 {
		panic(fmt.Sprintf("border thickness %d < 0", thick))
	}
	return func(o *borderOptions) { o.border.left = thick }
}

// PaddingAll option sets the padding size (in pixels) on all sides.
func PaddingAll(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("padding %d < 0", px))
	}
	return func(o *borderOptions) { o.padding = sides{px, px, px, px} }
}

// PaddingTop option sets the top padding size, in pixels.
func PaddingTop(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("padding %d < 0", px))
	}
	return func(o *borderOptions) { o.padding.top = px }
}

// PaddingRight option sets the right padding size, in pixels.
func PaddingRight(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("padding %d < 0", px))
	}
	return func(o *borderOptions) { o.padding.right = px }
}

// PaddingBottom option sets the bottom padding size, in pixels.
func PaddingBottom(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("padding %d < 0", px))
	}
	return func(o *borderOptions) { o.padding.bottom = px }
}

// PaddingLeft option sets the left padding size, in pixels.
func PaddingLeft(px int) BorderOption {
	if px < 0 {
		panic(fmt.Sprintf("padding %d < 0", px))
	}
	return func(o *borderOptions) { o.padding.left = px }
}

func evalBorderOptions(opts ...BorderOption) borderOptions {
	o := borderOptions{}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// Events implements the Env interface.
func (b Border) Events() <-chan gui.Event { return b.events }

// Draw implements the Env interface.
func (b Border) Draw() chan<- func(draw.Image) image.Rectangle { return b.draw }
