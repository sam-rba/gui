package gui

import (
	"image"
	"image/draw"

	"github.com/faiface/gui/lay/strain"
)

// Mux can be used to multiplex an Env, let's call it a root Env. Mux implements a way to
// create multiple virtual Envs that all interact with the root Env. They receive the same
// events, and their draw functions and constraints get redirected to the root Env.
type Mux struct {
	eventsIns chan chan<- Event
	draw      chan<- func(draw.Image) image.Rectangle
	impose    chan<- strain.Constraint
	finish    chan<- struct{}
}

// NewMux creates a new Mux that multiplexes the given Env. It returns the Mux along
// with a master Env. The master Env is just like any other Env created by the Mux,
// except that closing the master Env closes the whole Mux and all other Envs created
// by the Mux.
func NewMux(env Env) (mux *Mux, master Env) {
	finish := make(chan struct{})
	mux = &Mux{
		make(chan chan<- Event),
		env.Draw(),
		env.Impose(),
		finish,
	}

	go func() {
		var (
			eventsIns  []chan<- Event // one per child Env
			lastResize Event          // TODO: should we block until we receive the first one from the root Env?
		)
		defer func() {
			for _, evIn := range eventsIns {
				close(evIn)
			}
			close(mux.eventsIns)
			close(mux.finish)
			env.Close()
		}()
		for {
			select {
			case e := <-env.Events():
				if resize, ok := e.(Resize); ok {
					lastResize = resize
				}
				for _, eventsIn := range eventsIns {
					eventsIn <- e
				}
			case evIn := <-mux.eventsIns: // new env created by makeEnv()
				eventsIns = append(eventsIns, evIn)
				// Make sure to always send a resize event to a new Env if we got the size already.
				if lastResize != nil {
					evIn <- lastResize
				}
			case <-finish:
				return
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
	impose chan<- strain.Constraint
	finish chan<- struct{}
}

func (m *muxEnv) Events() <-chan Event                          { return m.events }
func (m *muxEnv) Draw() chan<- func(draw.Image) image.Rectangle { return m.draw }
func (m *muxEnv) Impose() chan<- strain.Constraint              { return m.impose }

func (m *muxEnv) Close() {
	m.finish <- *new(struct{})
	close(m.finish)
	close(m.draw)
	close(m.impose)
}

func (mux *Mux) makeEnv(master bool) Env {
	eventsOut, eventsIn := MakeEventsChan()
	draws := make(chan func(draw.Image) image.Rectangle)
	imposes := make(chan strain.Constraint)
	finish := make(chan struct{})
	env := &muxEnv{eventsOut, draws, imposes, finish}

	mux.eventsIns <- eventsIn

	go func() {
		if master {
			// Close Mux and all other child Envs when master Env is closed
			defer func() { mux.finish <- *new(struct{}) }()
		}

		// When the master Env gets closed, the Mux closes all the Events()
		// channels of all the children Envs, and it also closes the Mux's draw channel. However,
		// some child Envs of the Mux may still send some drawing commmands before they realize
		// that their Events() channel got closed.
		//
		// That is perfectly fine if their drawing commands simply get ignored. This down here
		// is a little hacky, but (I hope) perfectly fine solution to the problem.
		//
		// When the Mux's draw channel gets closed, the line marked with ! will
		// cause panic. We recover this panic, then we receive, but ignore all further draw
		// commands, correctly draining the Env until it closes itself.
		defer func() {
			if recover() != nil {
				for range draws {
				}
			}
		}()
		for {
			select {
			case d := <-draws:
				mux.draw <- d // !
			case c := <-imposes:
				mux.impose <- c
			case <-finish:
				return
			}
		}
	}()

	return env
}
