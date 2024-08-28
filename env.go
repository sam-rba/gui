package gui

import (
	"image"
	"image/draw"

	"git.samanthony.xyz/share"
)

// Env is the most important thing in this package. It is an interactive graphical
// environment, such as a window.
//
// The Events() channel produces events, like mouse and keyboard presses, while the
// Draw() channel receives drawing functions. A drawing function draws onto the
// supplied draw.Image, which is the drawing area of the Env and returns a rectangle
// covering the whole part of the image that got changed.
//
// An Env guarantees to produce a "resize/<x0>/<y0>/<x1>/<y1>" event as its first event.
//
// The Events() channel must be unlimited in capacity. Use share.Queue to create
// a channel of events with an unlimited capacity.
//
// The Draw() channel may be synchronous.
//
// Drawing functions sent to the Draw() channel are not guaranteed to be executed.
type Env interface {
	Events() <-chan Event
	Draw() chan<- func(draw.Image) image.Rectangle
	Killable

	killer
}

type env struct {
	events     <-chan Event
	draw       chan<- func(draw.Image) image.Rectangle
	attachChan chan<- victim
	kill       chan<- bool
	dead       <-chan bool
	detachChan <-chan bool
}

// newEnv makes an Env that receives Events from, and sends draw functions to, the parent.
//
// Each Event received from parent is passed to filterEvents() along with the Events() channel
// of the Env. Each draw function received from the Env's Draw() channel is passed to
// filterDraws() along with the Draw() channel of the parent Env.
//
// filterEvents() and filterDraws() can be used to, e.g., simply forward the Event or draw function
// to the channel, modify it before sending, not send it at all, or produce some other side-effects.
//
// shutdown() is called before the Env dies.
func newEnv(parent Env,
	filterEvents func(Event, chan<- Event),
	filterDraws func(func(draw.Image) image.Rectangle, chan<- func(draw.Image) image.Rectangle),
	shutdown func(),
) Env {
	events := share.NewQueue[Event]()
	drawChan := make(chan func(draw.Image) image.Rectangle)
	child := newKiller()
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
		defer shutdown()
		defer close(events.Enqueue)
		defer close(drawChan)
		defer close(kill)
		defer func() {
			go drain(drawChan)
			child.Kill() <- true
			<-child.Dead()
		}()

		for {
			select {
			case e := <-parent.Events():
				filterEvents(e, events.Enqueue)
			case d := <-drawChan:
				filterDraws(d, parent.Draw())
			case <-kill:
				return
			}
		}
	}()

	e := env{
		events:     events.Dequeue,
		draw:       drawChan,
		attachChan: child.attach(),
		kill:       kill,
		dead:       dead,
		detachChan: detachFromParent,
	}
	parent.attach() <- e
	return e
}

func (e env) Events() <-chan Event {
	return e.events
}

func (e env) Draw() chan<- func(draw.Image) image.Rectangle {
	return e.draw
}

func (e env) Kill() chan<- bool {
	return e.kill
}

func (e env) Dead() <-chan bool {
	return e.dead
}

func (e env) attach() chan<- victim {
	return e.attachChan
}

func (e env) detach() <-chan bool {
	return e.detachChan
}

func send[T any](v T, c chan<- T) {
	c <- v
}
