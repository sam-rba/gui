package gui

import (
	"image"
	"image/draw"
	"time"

	"git.samanthony.xyz/share"
)

const timeout = 1 * time.Second

// trySend returns true if v can be sent to c within timeout, or false otherwise.
func trySend[T any](c chan<- T, v T, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	select {
	case c <- v:
		return true
	case <-timer.C:
		return false
	}
}

// tryRecv returns the value received from c, or false if no value is received within timeout.
func tryRecv[T any](c <-chan T, timeout time.Duration) (*T, bool) {
	timer := time.NewTimer(timeout)
	select {
	case v := <-c:
		return &v, true
	case <-timer.C:
		return nil, false
	}
}

type dummyEnv struct {
	events share.Queue[Event]

	drawIn  chan<- func(draw.Image) image.Rectangle
	drawOut <-chan func(draw.Image) image.Rectangle

	kill chan<- bool
	dead <-chan bool

	attachChan chan<- victim
}

func newDummyEnv(size image.Rectangle) dummyEnv {
	events := share.NewQueue[Event]()
	drawIn := make(chan func(draw.Image) image.Rectangle)
	drawOut := make(chan func(draw.Image) image.Rectangle)
	kill := make(chan bool)
	dead := make(chan bool)

	child := newKiller()

	go func() {
		defer func() {
			dead <- true
			close(dead)
		}()
		defer close(kill)
		defer close(drawOut)
		defer close(drawIn)
		defer close(events.Enqueue)
		defer func() {
			go drain(drawIn)
			child.Kill() <- true
			<-child.Dead()
		}()

		for {
			select {
			case d := <-drawIn:
				drawOut <- d
			case <-kill:
				return
			}
		}
	}()

	events.Enqueue <- Resize{size}

	return dummyEnv{events, drawIn, drawOut, kill, dead, child.attach()}
}

func (de dummyEnv) Events() <-chan Event {
	return de.events.Dequeue
}

func (de dummyEnv) Draw() chan<- func(draw.Image) image.Rectangle {
	return de.drawIn
}

func (de dummyEnv) Kill() chan<- bool {
	return de.kill
}

func (de dummyEnv) Dead() <-chan bool {
	return de.dead
}

func (de dummyEnv) attach() chan<- victim {
	return de.attachChan
}

type dummyEvent struct {
	s string
}

func (e dummyEvent) String() string {
	return e.s
}
