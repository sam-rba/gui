package main

import (
	"image"
	"image/draw"

	"github.com/faiface/gui"
)

type envPair struct {
	env    gui.Env
	events <-chan gui.Event
	draw   chan<- func(draw.Image) image.Rectangle
	impose chan<- gui.Constraint
}

func (ep *envPair) Events() <-chan gui.Event                      { return ep.events }
func (ep *envPair) Draw() chan<- func(draw.Image) image.Rectangle { return ep.draw }
func (ep *envPair) Impose() chan<- gui.Constraint                 { return ep.impose }
func (ep *envPair) Close()                                        { ep.env.Close() }

func FixedLeft(env gui.Env, maxX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Max.X = maxX
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{env, out, env.Draw(), env.Impose()}
}

func FixedRight(env gui.Env, minX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Min.X = minX
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{env, out, env.Draw(), env.Impose()}
}

func FixedTop(env gui.Env, maxY int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Max.Y = maxY
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{env, out, env.Draw(), env.Impose()}
}

func FixedBottom(env gui.Env, minY int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Min.Y = minY
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{env, out, env.Draw(), env.Impose()}
}

func EvenHorizontal(env gui.Env, minI, maxI, n int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				x0, x1 := resize.Min.X, resize.Max.X
				resize.Min.X, resize.Max.X = x0+(x1-x0)*minI/n, x0+(x1-x0)*maxI/n
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{env, out, env.Draw(), env.Impose()}
}
