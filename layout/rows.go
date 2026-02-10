package layout

import (
	"image"
	"image/draw"

	"github.com/faiface/gui"
)

// NewRows creates layout with nrows children arranged in rows.
// It returns a slice containing an Env for each row.
// The height of each row is determined by the draw calls received from that row.
func NewRows(env gui.Env, nrows uint, o ...Option) []gui.Env {
	opts := evalOptions(o...)

	// Create event and draw channels for each row
	eventss := make([]chan gui.Event, nrows)                       // to children
	drawss := make([]chan func(draw.Image) image.Rectangle, nrows) // from children
	var i uint
	for i = 0; i < nrows; i++ {
		eventss[i] = make(chan gui.Event)
		drawss[i] = make(chan func(draw.Image) image.Rectangle)
	}

	go func() {
		defer close(env.Draw())
		defer closeAll(eventss)

		resize := func(area image.Rectangle, rowHeights []uint) {
			// Redraw background
			env.Draw() <- drawSubImage(drawBackground(opts.bg), area)

			// Send resize events to rows
			off := area.Min // vertical offset from parent area origin
			var i uint
			for i = 0; i < nrows; i++ {
				eventss[i] <- gui.Resize{image.Rectangle{
					off,
					off.Add(image.Pt(area.Dx(), int(rowHeights[i])))}}
				off.Y += int(rowHeights[i])
			}
		}

		// Receive and send first Resize event
		event := (<-env.Events()).(gui.Resize) // first event guaranteed to be Resize
		area := event.Rectangle
		rowHeights := make([]uint, nrows) // initially zero until draw call received
		resize(area, rowHeights)          // send first Resize to children

		// Multiplex rows' draw channels. Tag draw functions with row index.
		draws := make(chan taggedDrawCall)
		wg := newWaitGroup(nrows) // done when all rows close their Draw channel
		var i uint
		for i = 0; i < nrows; i++ {
			go muxDrawCalls(drawss[i], i, draws, wg)
		}
		defer close(draws)

		for {
			select {
			case event := <-env.Events(): // event from parent
				switch event := event.(type) {
				case gui.Resize:
					area = event.Rectangle
					resize(area, rowHeights)
				default:
					multicast(event, eventss) // forward event to all rows
				}
			case drw := <-draws: // draw call from a row
				rh := rowHeight(area, drw.f)
				oldrh := rowHeights[drw.idx]
				rowHeights[drw.idx] = rh
				if rh != oldrh { // size changed; redraw all rows
					go resize(area, rowHeights)
				} else { // Same size; just redraw the one row
					env.Draw() <- drawSubImage(drw.f, rowArea(area, rowHeights, drw.idx))
				}
			case <-wg.Wait(): // all rows' draw channels closed
				return
			}
		}
	}()

	// Create and return row Envs
	rows := make([]gui.Env, nrows)
	for i := range rows {
		rows[i] = rowEnv{eventss[i], drawss[i]}
	}
	return rows
}

// RowHeight calculates the height of a row within the area of the layout
// using a draw function received from the row.
func rowHeight(area image.Rectangle, drw func(draw.Image) image.Rectangle) uint {
	img := image.NewAlpha(area)
	return uint(drw(img).Canon().Dy())
}

// RowArea returns the drawing area of row i within the area of the layout.
func rowArea(area image.Rectangle, rowHeights []uint, i uint) image.Rectangle {
	min := area.Min.Add(image.Pt(0, int(sum(rowHeights[:i]))))
	max := min.Add(image.Pt(area.Dx(), int(rowHeights[i])))
	return image.Rectangle{min, max}
}

type rowEnv struct {
	events <-chan gui.Event
	draw   chan<- func(draw.Image) image.Rectangle
}

// Events implements the Env interface.
func (r rowEnv) Events() <-chan gui.Event { return r.events }

// Draw implements the Env interface.
func (r rowEnv) Draw() chan<- func(draw.Image) image.Rectangle { return r.draw }

func closeAll[T any](cs []chan T) {
	for _, c := range cs {
		close(c)
	}
}
