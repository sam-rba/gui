package gui

import (
	"image"
	"image/draw"
)

// Env is the most important thing in this package. It is an interactive graphical
// environment, such as a window.
type Env interface {
	// The Events() channel produces events, like mouse and keyboard presses.
	//
	// An Env guarantees to produce a "resize/<x0>/<y0>/<x1>/<y1>" event as its first event.
	//
	// The Events() channel must be unlimited in capacity. Use MakeEventsChan() to create
	// a channel of events with an unlimited capacity.
	Events() <-chan Event

	// The Draw() channel receives drawing functions.
	//
	// A drawing function draws onto the supplied draw.Image, which is the drawing area
	// of the Env, and returns a rectangle covering the whole part of the image that
	// got changed.
	//
	// Drawing functions are not guaranteed to be executed.
	//
	// The Draw() channel may be synchronous.
	Draw() chan<- func(draw.Image) image.Rectangle

	// The Impose() channel receives constraints that are imposed on the size/layout of
	// the Env by the widget occupying it.
	//
	// The Env may respond to constraints by sending a Resize event on the Events() channel.
	// However, constraints are not guaranteed to be satisfied, e.g. if there is not
	// enough space.
	//
	// The Impose() channel may be synchronous.
	Impose() chan<- Constraint

	// Close destroys the Env. The Env will subsequently close the Events(), Draw(),
	// and Impose() channels.
	Close()
}
