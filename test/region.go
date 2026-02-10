package main

import (
	"image"
	"image/draw"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/layout"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
)

var (
	bg = gui.HexToColor("#999999") // background color
)

func main() {
	mainthread.Run(run)
}

func run() {
	w, err := win.New(win.Title("Grid Layout Test"), win.Resizable())
	if err != nil {
		panic(err)
	}

	mux, env := gui.NewMux(w)

	// Create region in bottom-right quadrant of window
	resize := func(r image.Rectangle) image.Rectangle {
		return image.Rect(r.Min.X+r.Dx()/2, r.Min.Y+r.Dy()/2, r.Max.X, r.Max.Y)
	}
	region := layout.NewRegion(mux.MakeEnv(), resize, layout.Background(bg))
	go blinker(region)

	for event := range env.Events() {
		switch event.(type) {
		case win.WiClose:
			close(env.Draw())
			return
		}
	}
}

// Blinker is a widget that blinks three times when it is clicked.
func blinker(env gui.Env) {
	redraw := func(visible bool) func(draw.Image) image.Rectangle {
		return func(img draw.Image) image.Rectangle {
			if visible {
				draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)
			} else {
				draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
			}
			return img.Bounds()
		}
	}

	area := (<-env.Events()).(gui.Resize).Rectangle // first event guaranteed to be Resize
	env.Draw() <- redraw(true)

	for event := range env.Events() {
		switch event := event.(type) {
		case gui.Resize:
			area = event.Rectangle
			env.Draw() <- redraw(true)
		case win.MoDown:
			if event.Point.In(area) {
				go func() {
					for i := 0; i < 3; i++ {
						env.Draw() <- redraw(false)
						time.Sleep(time.Second / 6)
						env.Draw() <- redraw(true)
						time.Sleep(time.Second / 6)
					}
				}()
			}
		}
	}
	close(env.Draw())
}
