package layout

import (
	"image"
	"image/draw"

	"github.com/faiface/gui"
)

// Region is a layout with a single child Env that occupies a sub-area of its parent.
type Region struct {
	events <-chan gui.Event
	draw   chan<- func(draw.Image) image.Rectangle
}

// NewRegion creates a region layout that occupies part of the parent env's area, as determined by the resize function.
// Resize takes the area of the parent and returns the area of the region.
func NewRegion(env gui.Env, resize func(image.Rectangle) image.Rectangle) gui.Env {
	events := make(chan gui.Event)                     // to child
	drw := make(chan func(draw.Image) image.Rectangle) // from child

	go func(events chan<- gui.Event, drw <-chan func(draw.Image) image.Rectangle) {
		// Forward first resize event to child
		event := <-env.Events()
		area := resize(event.(gui.Resize).Rectangle) // first event guaranteed to be Resize
		events <- gui.Resize{area}

		for {
			select {
			case event := <-env.Events():
				switch event := event.(type) {
				case gui.Resize:
					area = resize(event.Rectangle)
					events <- gui.Resize{area}
				default:
					events <- event
				}
			case f, ok := <-drw:
				if !ok {
					close(events)
					close(env.Draw())
					return
				}
				env.Draw() <- drawRegion(f, area)
			}
		}
	}(events, drw)

	return Region{events, drw}
}

// Translate a draw call to the given area.
func drawRegion(f func(draw.Image) image.Rectangle, area image.Rectangle) func(draw.Image) image.Rectangle {
	return func(img draw.Image) image.Rectangle {
		return f(subimage(img, area))
	}
}

// Events implements the Env interface.
func (r Region) Events() <-chan gui.Event { return r.events }

// Draw implements the Env interface.
func (r Region) Draw() chan<- func(draw.Image) image.Rectangle { return r.draw }
