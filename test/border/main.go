package main

import (
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/layout"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
)

const (
	margin  = 10
	border  = 2
	padding = 15
)

var (
	bgClr   = gui.HexToColor("#999999") // background color
	brdrClr = color.RGBA{0xFF, 0x00, 0xFF, 0xFF}
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

	// Background
	bg := layout.NewRegion(mux.MakeEnv(), bgClr, layout.Full())

	// Margin, border, and padding
	border := layout.NewBorder(bg, layout.Margin(margin), layout.Border(border, brdrClr), layout.Padding(padding))

	//region := layout.NewRegion(border, color.White, layout.Full())
	//go func() { for range region.Events() {} }()

	// Region in top-right quadrant
	region := layout.NewRegion(border, color.Transparent, layout.Quad1())
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
