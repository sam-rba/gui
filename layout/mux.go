package layout

import (
	"image"
	"image/draw"
	"log"

	"git.samanthony.xyz/share"
	"github.com/faiface/gui"
)

// Mux can be used to multiplex an Env, let's call it a root Env. Mux implements a way to
// create multiple virtual Envs that all interact with the root Env. They receive the same
// events apart from gui.Resize, and their draw functions get redirected to the root Env.
//
// All gui.Resize events are instead modified according to the underlying Layout.
type Mux struct {
	// Sending any value to Kill will terminate the Mux.
	Kill chan<- any

	bounds    share.Val[image.Rectangle]
	draw      chan<- func(draw.Image) image.Rectangle
	eventsIns share.ConstSlice[chan<- gui.Event]
	layout    Layout
}

// Layout returns the underlying Layout of the Mux.
func (mux *Mux) Layout() Layout {
	return mux.layout
}

func NewMux(parent gui.Env, children []*gui.Env, layout Layout) Mux {
	parent = layout.Intercept(parent)

	kill := make(chan any)
	bounds := share.NewVal[image.Rectangle]()
	drawChan := make(chan func(draw.Image) image.Rectangle)
	eventsIns := func() share.ConstSlice[chan<- gui.Event] { // create child Env's
		evIns := make([]chan<- gui.Event, len(children))
		for i, child := range children {
			*child, evIns[i] = makeEnv(drawChan)
		}
		return share.NewConstSlice(evIns)
	}()
	mux := Mux{
		kill,
		bounds,
		drawChan,
		eventsIns,
		layout,
	}

	go func() {
		defer close(parent.Draw())
		defer close(kill)
		defer bounds.Close()
		defer close(drawChan)
		defer func() {
			for eventsIn := range eventsIns.Elems() {
				close(eventsIn)
			}
			eventsIns.Close()
		}()

		for {
			select {
			case <-kill:
				return
			case d := <-drawChan:
				parent.Draw() <- d
			case e := <-parent.Events():
				if resize, ok := e.(gui.Resize); ok {
					bounds.Set <- resize.Rectangle
					mux.resizeChildren()
				} else {
					for eventsIn := range eventsIns.Elems() {
						eventsIn <- e
					}
				}
			}
		}
	}()

	// First event of a new Env must be Resize.
	mux.resizeChildren()

	return mux
}

func (mux *Mux) resizeChildren() {
	rect := mux.bounds.Get()
	lay := mux.layout.Lay(rect)
	i := 0
	for eventsIn := range mux.eventsIns.Elems() {
		if i > len(lay) {
			log.Printf("Lay of %T is not large enough (%d) for the number of children, skipping\n",
				mux.layout, len(lay))
			break
		}
		eventsIn <- gui.Resize{lay[i]}
		i++
	}
}

type muxEnv struct {
	events <-chan gui.Event
	draw   chan<- func(draw.Image) image.Rectangle
}

func (m *muxEnv) Events() <-chan gui.Event                      { return m.events }
func (m *muxEnv) Draw() chan<- func(draw.Image) image.Rectangle { return m.draw }

func makeEnv(muxDraw chan<- func(draw.Image) image.Rectangle) (env gui.Env, eventsIn chan<- gui.Event) {
	eventsOut, eventsIn := gui.MakeEventsChan()
	envDraw := make(chan func(draw.Image) image.Rectangle)
	env = &muxEnv{eventsOut, envDraw}

	go func() {
		func() {
			// When the master Env gets its Draw() channel closed, it closes all the Events()
			// channels of all the children Envs, and it also closes the internal draw channel
			// of the Mux. Otherwise, closing the Draw() channel of the master Env wouldn't
			// close the Env the Mux is muxing. However, some child Envs of the Mux may still
			// send some drawing commmands before they realize that their Events() channel got
			// closed.
			//
			// That is perfectly fine if their drawing commands simply get ignored. This down here
			// is a little hacky, but (I hope) perfectly fine solution to the problem.
			//
			// When the internal draw channel of the Mux gets closed, the line marked with ! will
			// cause panic. We recover this panic, then we receive, but ignore all furhter draw
			// commands, correctly draining the Env until it closes itself.
			defer func() {
				if recover() != nil {
					for range envDraw {
					}
				}
			}()
			for d := range envDraw {
				muxDraw <- d // !
			}
		}()
	}()

	return env, eventsIn
}
