package gui

import "image"

// Scheme represents the appearance and behavior of a layout.
type Scheme interface {
	// The Partitioner represents the way to divide space among the children.
	// It takes a parameter of how much space is available, and returns a space for each child.
	Partitioner

	// The Intercepter transforms an Env channel to another.
	// This way the Layout can emit its own Events, re-emit previous ones,
	// or even stop an event from propagating, think win.MoScroll.
	// It can be a no-op.
	Intercepter
}

// Partitioner divides a large Rectangle into several smaller sub-Rectangles.
type Partitioner interface {
	Partition(image.Rectangle) []image.Rectangle
}

// NewLayout takes an array of uninitialized `child' Envs and multiplexes the `parent' Env
// according to the provided Scheme. The children receive the same events from the parent
// aside from Resize, and their draw functions get redirected to the parent Env.
//
// The Scheme determines the look and behavior of the Layout. Resize events for each child
// are modified according to the Partitioner. Other Events and draw functions can be modified
// by the Intercepter.
//
// Killing the returned layout kills all of the children.
func NewLayout(parent Env, children []*Env, scheme Scheme) Killable {
	env := newEnv(parent, send, send, func() {})

	// Capture Resize Events to be sent to the Partitioner.
	resizeSniffer, resizes := newSniffer(env, func(e Event) (r image.Rectangle, ok bool) {
		if resize, ok := e.(Resize); ok {
			return resize.Rectangle, true
		}
		return image.Rectangle{}, false
	})

	intercepter := scheme.Intercept(resizeSniffer)

	mux := NewMux(intercepter)
	muxEnvs := make([]Env, len(children))
	resizers := make([]Env, len(children))
	resizerChans := make([]chan image.Rectangle, len(children))
	for i, child := range children {
		muxEnvs[i] = mux.MakeEnv()
		resizerChans[i] = make(chan image.Rectangle)
		resizers[i] = newResizer(muxEnvs[i], resizerChans[i])
		*child = resizers[i]
	}

	go func() {
		for rect := range resizes {
			for i, r := range scheme.Partition(rect) {
				resizerChans[i] <- r
			}
		}
		for _, c := range resizerChans {
			close(c)
		}
	}()

	return env
}

// newSniffer makes an Env that forwards all Events and Draws unchanged, but emits a signal
// whenever a certain event is encountered. It returns the new Env and the signal channel.
//
// Each Event from parent is passed to sniff(). If sniff() accepts the Event, the return
// value of sniff() is sent to the signal channel.
//
// signal is closed automatically when the sniffer dies.
func newSniffer[T any](parent Env, sniff func(Event) (v T, ok bool)) (Env, <-chan T) {
	signal := make(chan T)
	env := newEnv(parent,
		func(e Event, c chan<- Event) {
			c <- e
			if sig, ok := sniff(e); ok {
				signal <- sig
			}
		},
		send, // forward draw functions un-modified
		func() {
			close(signal)
		})
	return env, signal
}

// newResizer makes an Env that replaces the values all Resize events with
// Rectangles received from the resize channel.
// It waits for a Rectangle from resize each time a Resize Event is received from parent.
func newResizer(parent Env, resize <-chan image.Rectangle) Env {
	return newEnv(parent,
		func(e Event, c chan<- Event) {
			if _, ok := e.(Resize); ok {
				e = Resize{<-resize}
			}
			c <- e
		},
		send, // forward draw functions un-modified
		func() {})
}
