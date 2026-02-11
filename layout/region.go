package layout

import (
	"image"
	"image/color"
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
// It returns the child Env.
func NewRegion(env gui.Env, clr color.Color, resize ResizeFunc) gui.Env {
	events := make(chan gui.Event)                       // to child
	draws := make(chan func(draw.Image) image.Rectangle) // from child

	go func(events chan<- gui.Event, draws <-chan func(draw.Image) image.Rectangle) {
		defer close(env.Draw())
		defer close(events)

		redraw := func(area, childArea image.Rectangle) {
			env.Draw() <- drawSubImage(drawBackground(clr), area)
			events <- gui.Resize{childArea}
		}

		// Draw background and forward first resize event to child
		event := (<-env.Events()).(gui.Resize) // first event guaranteed to be Resize
		area := event.Rectangle
		childArea := resize(area)
		go redraw(area, childArea)

		for {
			select {
			case event := <-env.Events(): // event from parent
				switch event := event.(type) {
				case gui.Resize:
					area = event.Rectangle
					childArea = resize(area)
					go redraw(area, childArea)
				default:
					go func() { events <- event }() // forward to child
				}
			case f, ok := <-draws: // draw call from child
				if !ok {
					return
				}
				env.Draw() <- drawSubImage(f, childArea)
			}
		}
	}(events, draws)

	return Region{events, draws}
}

// Events implements the Env interface.
func (r Region) Events() <-chan gui.Event { return r.events }

// Draw implements the Env interface.
func (r Region) Draw() chan<- func(draw.Image) image.Rectangle { return r.draw }
