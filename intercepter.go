package gui

import (
	"image"
	"image/draw"
)

// Intercepter represents an element that can interact with Envs.
// An Intercepter can modify Events, stop them or emit arbitrary ones.
// It can also put itself in the draw pipeline, for throttling very
// expensive draw calls for example.
type Intercepter interface {
	Intercept(Env) Env
}

// RedrawIntercepter is a basic Intercepter, it is meant for use in simple Layouts
// that only need to redraw themselves.
type RedrawIntercepter struct {
	Redraw func(draw.Image, image.Rectangle)
}

func (ri RedrawIntercepter) Intercept(parent Env) Env {
	return newEnv(parent,
		func(e Event, c chan<- Event) {
			c <- e
			if resize, ok := e.(Resize); ok {
				parent.Draw() <- func(drw draw.Image) image.Rectangle {
					ri.Redraw(drw, resize.Rectangle)
					return resize.Rectangle
				}
			}
		},
		send, // forward draw functions un-modified
		func() {})
}
