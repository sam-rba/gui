package gui

import (
	"image"
	"image/draw"
	"image/jpeg"
	"os"
	"testing"
	"time"

	"github.com/fogleman/gg"
)

// Send Events from the Mux to the Envs..
func TestMuxEvent(t *testing.T) {
	rect := image.Rect(12, 34, 56, 78)
	root := newDummyEnv(rect)
	defer func() {
		root.Kill() <- true
		<-root.Dead()
	}()
	mux := NewMux(root)
	envs := []Env{mux.MakeEnv(), mux.MakeEnv(), mux.MakeEnv()}

	events := []Event{Resize{rect}, dummyEvent{"fooEvent"}, dummyEvent{"barEvent"}, dummyEvent{"bazEvent"}}
	go func() {
		for _, event := range events[1:] { // skip resize—it's sent automatically by the root Env
			root.eventsIn <- event
		}
	}()

	for _, env := range envs {
		for _, expect := range events {
			var event Event
			timer := time.NewTimer(timeout)
			select {
			case event = <-env.Events():
			case <-timer.C:
				t.Errorf("no event received after %v", timeout)
			}
			if event != expect {
				t.Errorf("received %v; wanted %v", event, expect)
			}
		}
	}
}

// Send draw function from Envs to the Mux.
func TestMuxDraw(t *testing.T) {
	rect := image.Rect(120, 340, 560, 780)
	root := newDummyEnv(rect)
	defer func() {
		root.Kill() <- true
		<-root.Dead()
	}()
	mux := NewMux(root)
	envs := []Env{mux.MakeEnv(), mux.MakeEnv(), mux.MakeEnv()}
	drawFunc := func(r image.Rectangle) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			cr := image.Rect(0, 0, r.Dx(), r.Dy())
			canvas := image.NewRGBA(cr)
			draw.Draw(canvas, cr, image.White, image.ZP, draw.Src)
			dc := gg.NewContextForRGBA(canvas)
			dc.DrawEllipse(float64(r.Dx()/2), float64(r.Dy()/2),
				float64(r.Dx()/2), float64(r.Dy()/2))
			dc.SetRGB(255, 120, 0)
			dc.Fill()
			draw.Draw(drw, r, canvas, image.ZP, draw.Src)
			return r
		}
	}

	for _, env := range envs {
		// Receive Resize event.
		var event Event
		timer := time.NewTimer(timeout)
		select {
		case event = <-env.Events():
		case <-timer.C:
			t.Errorf("no event received after %v", timeout)
		}
		resize, ok := event.(Resize)
		if !ok {
			t.Errorf("got %T; wanted Resize", event)
		}

		env.Draw() <- drawFunc(resize.Rectangle)

		// Receive draw function on root env.
		var d func(draw.Image) image.Rectangle
		timer.Reset(timeout)
		select {
		case d = <-root.drawOut:
		case <-timer.C:
			t.Errorf("no draw function received after %v", timeout)
		}

		// Draw and compare images.
		img := image.NewRGBA(resize.Rectangle)
		r := d(img)
		expectImg := image.NewRGBA(rect)
		expectR := drawFunc(rect)(expectImg)
		if r != expectR {
			t.Errorf("draw function returned %v; wanted %v", r, expectR)
		}
		if !cmpImg(img, expectImg) {
			expectOut, gotOut := "expect.jpg", "got.jpg"
			t.Errorf("draw function did not draw correct image. Writing results to '%v' and '%v'...", expectOut, gotOut)
			if err := writeImg(expectImg, expectOut); err != nil {
				t.Error(err)
			}
			if err := writeImg(img, gotOut); err != nil {
				t.Error(err)
			}
		}
	}
}

func cmpImg(a, b image.Image) bool {
	if a.Bounds() != b.Bounds() {
		return false
	}
	bounds := a.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}
	return true
}

func writeImg(img image.Image, fname string) error {
	f, err := os.Create(fname)
	defer f.Close()
	if err != nil {
		return err
	}
	return jpeg.Encode(f, img, nil)
}

type dummyEnv struct {
	eventsIn  chan<- Event
	eventsOut <-chan Event

	drawIn  chan<- func(draw.Image) image.Rectangle
	drawOut <-chan func(draw.Image) image.Rectangle

	kill chan<- bool
	dead <-chan bool

	attachChan chan<- attachable
}

func newDummyEnv(size image.Rectangle) dummyEnv {
	eventsOut, eventsIn := MakeEventsChan()
	drawIn := make(chan func(draw.Image) image.Rectangle)
	drawOut := make(chan func(draw.Image) image.Rectangle)
	kill := make(chan bool)
	dead := make(chan bool)

	attached := newAttachHandler()

	go func() {
		defer func() {
			dead <- true
			close(dead)
		}()
		defer close(kill)
		defer close(drawOut)
		defer close(drawIn)
		defer close(eventsIn)
		defer func() {
			attached.kill <- true
			<-attached.dead
		}()
		defer func() {
			go drain(drawIn)
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

	eventsIn <- Resize{size}

	return dummyEnv{eventsIn, eventsOut, drawIn, drawOut, kill, dead, attached.attach()}
}

func (de dummyEnv) Events() <-chan Event {
	return de.eventsOut
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

func (de dummyEnv) attach() chan<- attachable {
	return de.attachChan
}

type dummyEvent struct {
	s string
}

func (e dummyEvent) String() string {
	return e.s
}
