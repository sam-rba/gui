package gui

import (
	"fmt"
	"image"
	"image/draw"

	"git.samanthony.xyz/share"
)

// Mux can be used to multiplex an Env, let's call it the parent Env. Mux implements a way to
// create multiple virtual Envs that all interact with the parent Env. They receive the same
// events and their draw functions get redirected to the parent Env.
type Mux struct {
	size        share.Val[image.Rectangle]
	draw        chan<- func(draw.Image) image.Rectangle
	addChild    chan<- muxEnv
	removeChild chan<- muxEnv
	kill        chan<- bool
	dead        <-chan bool
	detachChan  <-chan bool
}

func NewMux(parent Env) Mux {
	size := share.NewVal[image.Rectangle]()
	drawChan := make(chan func(draw.Image) image.Rectangle)
	addChild := make(chan muxEnv)
	removeChild := make(chan muxEnv)
	kill := make(chan bool)
	dead := make(chan bool)

	detachFromParent := make(chan bool)

	go func() {
		defer func() {
			dead <- true
			close(dead)
		}()
		defer func() {
			detachFromParent <- true
			close(detachFromParent)
		}()
		defer close(kill)
		defer close(removeChild)
		defer close(addChild)
		defer close(drawChan)
		defer size.Close()

		var children []muxEnv
		defer func() {
			go drain(drawChan) // children may still be sending
			for _, child := range children {
				child.kill <- true
			}
			for range children {
				<-removeChild
			}
		}()

		for {
			select {
			case d := <-drawChan:
				parent.Draw() <- d
			case e := <-parent.Events():
				if resize, ok := e.(Resize); ok {
					size.Set <- resize.Rectangle
				}
				for _, child := range children {
					child.eventsIn <- e
				}
			case child := <-addChild:
				children = append(children, child)
			case child := <-removeChild:
				var err error
				// TODO: faster search
				if children, err = remove(child, children); err != nil {
					panic(fmt.Sprintf("Mux: failed to remove child Env: %v", err))
				}
			case <-kill:
				return
			}
		}
	}()

	mux := Mux{
		size:        size,
		draw:        drawChan,
		addChild:    addChild,
		removeChild: removeChild,
		kill:        kill,
		dead:        dead,
		detachChan:  detachFromParent,
	}
	parent.attach() <- mux
	return mux
}

func (mux Mux) Kill() chan<- bool {
	return mux.kill
}

func (mux Mux) Dead() <-chan bool {
	return mux.dead
}

func (mux Mux) detach() <-chan bool {
	return mux.detachChan
}

type muxEnv struct {
	eventsIn      chan<- Event
	eventsOut     <-chan Event
	draw          chan<- func(draw.Image) image.Rectangle
	attachChan    chan<- attachable
	kill          chan<- bool
	dead          <-chan bool
	detachFromMux <-chan bool
}

func (mux Mux) MakeEnv() Env {
	eventsOut, eventsIn := MakeEventsChan()
	drawChan := make(chan func(draw.Image) image.Rectangle)
	attached := newAttachHandler()
	kill := make(chan bool)
	dead := make(chan bool)
	detachFromMux := make(chan bool)

	env := muxEnv{
		eventsIn:      eventsIn,
		eventsOut:     eventsOut,
		draw:          drawChan,
		attachChan:    attached.attach(),
		kill:          kill,
		dead:          dead,
		detachFromMux: detachFromMux,
	}
	mux.addChild <- env
	// make sure to always send a resize event to a new Env
	eventsIn <- Resize{mux.size.Get()}

	go func() {
		defer func() {
			dead <- true
			close(dead)
		}()
		defer close(kill)
		defer close(drawChan)
		defer close(eventsIn)
		// eventsOut closed automatically by MakeEventsChan()

		defer func() {
			mux.removeChild <- env
		}()

		defer func() {
			attached.kill <- true
			<-attached.dead
		}()
		defer func() {
			go drain(drawChan)
		}()

		for {
			select {
			case d := <-drawChan:
				mux.draw <- d
			case <-kill:
				return
			}
		}
	}()

	return env
}

func (env muxEnv) Events() <-chan Event {
	return env.eventsOut
}

func (env muxEnv) Draw() chan<- func(draw.Image) image.Rectangle {
	return env.draw
}

func (env muxEnv) Kill() chan<- bool {
	return env.kill
}

func (env muxEnv) Dead() <-chan bool {
	return env.dead
}

func (env muxEnv) attach() chan<- attachable {
	return env.attachChan
}

// remove removes element e from slice s, returning the modified slice, or error if e is not in s.
func remove[S ~[]E, E comparable](e E, s S) ([]E, error) {
	for i := range s {
		if s[i] == e {
			return append(s[:i], s[i+1:]...), nil
		}
	}
	return s, fmt.Errorf("%v not found in %v", e, s)
}

func drain[T any](c <-chan T) {
	for range c {
	}
}
