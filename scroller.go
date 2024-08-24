package gui

import (
	"image"
	"image/color"
	"image/draw"

	"git.samanthony.xyz/share"
)

var _ Scheme = Scroller{}

type Scroller struct {
	Background  color.Color
	Length      int
	ChildHeight int
	Offset      int
	Gap         int
	Vertical    bool
}

func (s Scroller) redraw(drw draw.Image, bounds image.Rectangle) {
	col := s.Background
	if col == nil {
		col = image.Black
	}
	draw.Draw(drw, bounds, image.NewUniform(col), image.ZP, draw.Src)
}

func clamp(val, a, b int) int {
	if a > b {
		if val < b {
			return b
		}
		if val > a {
			return a
		}
	} else {
		if val > b {
			return b
		}
		if val < a {
			return a
		}
	}
	return val
}

func (s Scroller) Partition(bounds image.Rectangle) []image.Rectangle {
	items := s.Length
	ch := s.ChildHeight
	gap := s.Gap

	ret := make([]image.Rectangle, items)
	Y := bounds.Min.Y + s.Offset + gap
	for i := 0; i < items; i++ {
		r := image.Rect(bounds.Min.X+gap, Y, bounds.Max.X-gap, Y+ch)
		ret[i] = r
		Y += ch + gap
	}
	return ret
}

func (s Scroller) Intercept(parent Env) Env {
	lastResize := share.NewVal[image.Rectangle]()
	img := share.NewVal[draw.Image]()
	mouseOver := share.NewVal[bool]()

	lastResize.Set <- image.Rectangle{}
	img.Set <- image.NewRGBA(image.Rectangle{})
	mouseOver.Set <- false

	return newEnv(parent,
		func(event Event, events chan<- Event) {
			switch event := event.(type) {
			case MoMove:
				mouseOver.Set <- event.Point.In(lastResize.Get())
			case MoScroll:
				if !mouseOver.Get() {
					break
				}

				oldoff := s.Offset
				v := s.Length*s.ChildHeight + ((s.Length + 1) * s.Gap)
				bounds := lastResize.Get()

				if s.Vertical {
					h := bounds.Dx()
					s.Offset = clamp(s.Offset+event.Point.X*16, h-v, 0)
				} else {
					h := bounds.Dy()
					s.Offset = clamp(s.Offset+event.Point.Y*16, h-v, 0)
				}

				if oldoff != s.Offset {
					m := img.Get()
					s.redraw(m, m.Bounds())
					events <- Resize{bounds}
				}
			case Resize:
				lastResize.Set <- event.Rectangle

				m := image.NewRGBA(event.Rectangle)
				img.Set <- m
				s.redraw(m, m.Bounds())

				events <- event
			default:
				events <- event
			}
		},
		func(drawFunc func(draw.Image) image.Rectangle, drawChan chan<- func(draw.Image) image.Rectangle) {
			m := img.Get()
			if drawFunc(m).Intersect(m.Bounds()) != image.ZR {
				drawChan <- func(drw draw.Image) image.Rectangle {
					bounds := lastResize.Get()
					draw.Draw(drw, bounds, m, bounds.Min, draw.Over)
					return m.Bounds()
				}
			}
		},
		func() {
			lastResize.Close()
			img.Close()
			mouseOver.Close()
		})
}
