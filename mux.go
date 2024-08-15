package gui

import (
	"image"
	"image/draw"
)

// Mux can be used to multiplex an Env, let's call it a root Env. Mux implements a way to
// create multiple virtual Envs that all interact with the root Env. They receive the same
// events and their draw functions get redirected to the root Env.
type Mux struct {
	addEventsIn chan<- chan<- Event
	size sharedVal[image.Rectangle]
	draw chan<- func(draw.Image) image.Rectangle
}

// NewMux creates a new Mux that multiplexes the given Env. It returns the Mux along with
// a master Env. The master Env is just like any other Env created by the Mux, except that
// closing the Draw() channel on the master Env closes the whole Mux and all other Envs
// created by the Mux.
func NewMux(env Env) (mux *Mux, master Env) {
	addEventsIn := make(chan chan<- Event)
	size := newSharedVal[image.Rectangle]()
	drawChan := make(chan func(draw.Image) image.Rectangle)
	mux = &Mux{
		addEventsIn: addEventsIn,
		size:     size,
		draw:        drawChan,
	}

	go func() {
		var eventsIns []chan<- Event

		defer close(env.Draw())
		defer close(addEventsIn)
		defer size.close()
		defer func() {
			for _, eventsIn := range eventsIns {
				close(eventsIn)
			}
		}()

		for {
			select {
			case d, ok := <-drawChan:
				if !ok { // closed by master env
					return
				}
				env.Draw() <- d
			case e, ok := <-env.Events():
				if !ok {
					return
				}
				if resize, ok := e.(Resize); ok {
					size.set <- resize.Rectangle
				}
				for _, eventsIn := range eventsIns {
					eventsIn <- e
				}
			case eventsIn := <-addEventsIn:
				eventsIns = append(eventsIns, eventsIn)
			}
		}
	}()

	master = mux.makeEnv(true)
	return mux, master
}

// MakeEnv creates a new virtual Env that interacts with the root Env of the Mux. Closing
// the Draw() channel of the Env will not close the Mux, or any other Env created by the Mux
// but will delete the Env from the Mux.
func (mux *Mux) MakeEnv() Env {
	return mux.makeEnv(false)
}

type muxEnv struct {
	events <-chan Event
	draw   chan<- func(draw.Image) image.Rectangle
}

func (m *muxEnv) Events() <-chan Event                          { return m.events }
func (m *muxEnv) Draw() chan<- func(draw.Image) image.Rectangle { return m.draw }

func (mux *Mux) makeEnv(master bool) Env {
	eventsOut, eventsIn := MakeEventsChan()
	drawChan := make(chan func(draw.Image) image.Rectangle)
	env := &muxEnv{eventsOut, drawChan}

	mux.addEventsIn <- eventsIn
	// make sure to always send a resize event to a new Env
	eventsIn <- Resize{mux.size.get()}

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
					drain(drawChan)
				}
			}()
			for d := range drawChan {
				mux.draw <- d // !
			}
		}()
		if master {
			close(mux.draw)
		}
	}()

	return env
}

func drain[T any](c <-chan T) { for range c {} }
