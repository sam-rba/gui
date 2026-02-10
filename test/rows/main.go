package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
	"github.com/faiface/gui/layout"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
)

const (
	nrows     = 16
	rowHeight = 12
	rowWidth  = 128
)

var bgclr = gui.HexToColor("#ffffea") // background color

func main() {
	mainthread.Run(run)
}

func run() {
	w, err := win.New(win.Title("Grid Layout Test"), win.Resizable())
	if err != nil {
		panic(err)
	}

	mux, env := gui.NewMux(w)
	bg := layout.NewRegion(mux.MakeEnv(), bgclr, layout.ResizeAll)
	rows := layout.NewRows(bg, nrows)
	for i, row := range rows {
		go colorBlock(row, image.Pt(rowWidth, rowHeight), color.RGBA{uint8(i * 256 / 4), 0x20, 0x20, 0xFF})
	}

	for event := range env.Events() {
		switch event.(type) {
		case win.WiClose:
			close(env.Draw())
			return
		}
	}
}

func colorBlock(env gui.Env, size image.Point, c color.Color) {
	redraw := func(img draw.Image) image.Rectangle {
		r := image.Rectangle{img.Bounds().Min, img.Bounds().Min.Add(size)}
		draw.Draw(img, r, &image.Uniform{c}, image.ZP, draw.Src)
		return r
	}
	for event := range env.Events() {
		switch event.(type) {
		case gui.Resize:
			env.Draw() <- redraw
		}
	}
	close(env.Draw())
}
