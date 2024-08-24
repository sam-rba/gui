package gui

import (
	"image"
	"testing"
)

func TestSniffer(t *testing.T) {
	root := newDummyEnv(image.Rect(12, 34, 56, 78))
	defer func() {
		root.kill <- true
		<-root.dead
	}()

	expectSig := "got fooEvent"
	sniffer, signal := newSniffer(root, func(e Event) (string, bool) {
		if e.String() == "fooEvent" {
			return expectSig, true
		}
		return "", false
	})

	// First event should be Resize.
	eventp, ok := tryRecv(sniffer.Events(), timeout)
	if !ok {
		t.Fatalf("no Resize event received after %v", timeout)
	}
	if _, ok := (*eventp).(Resize); !ok {
		t.Fatalf("got %v Event; wanted Resize", *eventp)
	}

	for i := 0; i < 3; i++ { // arbitrary number of iterations
		// Send events to sniffer.
		events := []Event{dummyEvent{"barEvent"}, dummyEvent{"fooEvent"}}
		for _, event := range events {
			root.events.Enqueue <- event

			eventp, ok := tryRecv(sniffer.Events(), timeout)
			if !ok {
				t.Fatalf("no Event received after %v", timeout)
			}
			if *eventp != event {
				t.Errorf("received Event %v; wanted %v", *eventp, event)
			}
		}

		// One of the events should trigger a signal.
		sigp, ok := tryRecv(signal, timeout)
		if !ok {
			t.Fatalf("no signal received after %v", timeout)
		}
		if *sigp != expectSig {
			t.Errorf("received signal %v; wanted %v", *sigp, expectSig)
		}
	}
}

func TestResizer(t *testing.T) {
	root := newDummyEnv(image.Rectangle{})
	defer func() {
		root.kill <- true
		<-root.dead
	}()

	resizeChan := make(chan image.Rectangle)
	defer close(resizeChan)
	resizer := newResizer(root, resizeChan)

	sizes := []image.Rectangle{
		image.Rect(11, 22, 33, 44),
		image.Rect(55, 66, 77, 88),
		image.Rect(99, 111, 222, 333),
	}
	for _, size := range sizes {
		// First Resize event is sent automatically by root.

		if !trySend(resizeChan, size, timeout) {
			t.Errorf("resizer did not accept Rectangle after %v", timeout)
		}

		eventp, ok := tryRecv(resizer.Events(), timeout)
		if !ok {
			t.Fatalf("no Event received from resizer after %v", timeout)
		}
		resize, ok := (*eventp).(Resize)
		if !ok {
			t.Fatalf("received %v Event from resizer; wanted Resize", *eventp)
		}
		if resize.Rectangle != size {
			t.Errorf("got %v from resizer; wanted %v", resize.Rectangle, size)
		}

		// this event should be replaced by the resizer
		root.events.Enqueue <- Resize{image.Rectangle{}}
	}
}
