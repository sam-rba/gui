package style

import (
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

var defaultOptions = options{
	font: goregular.TTF,
	FaceOptions: opentype.FaceOptions{
		DefaultFontSize,
		DefaultDpi,
		DefaultHinting,
	},
}

// Option is a functional option to Style constructor, New.
type Option func(*options)

type options struct {
	font []byte
	opentype.FaceOptions
}

func parseOpts(opts ...Option) options {
	o := defaultOptions
	for i := range opts {
		opts[i](&o)
	}
	return o
}

// Font option sets the font source to use. It takes TTF or OTF data
// (see golang.org/x/image/font/opentype.Parse).
func Font(src []byte) Option {
	return func(o *options) {
		o.font = src
	}
}

// FontSize option sets the font size in points (1/72").
func FontSize(size float64) Option {
	return func(o *options) {
		o.FaceOptions.Size = size
	}
}

// Dpi option sets the dots per inch resolution of the display.
func Dpi(dpi float64) Option {
	return func(o *options) {
		o.FaceOptions.DPI = dpi
	}
}

// Hinting options selects how to quantize a vector font's glyph
// nodes.
func Hinting(h font.Hinting) Option {
	return func(o *options) {
		o.FaceOptions.Hinting = h
	}
}
