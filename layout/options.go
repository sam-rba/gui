package layout

import (
	"image/color"
)

var (
	defaultBg = color.Transparent
)

// Option is a function option for layout constructors.
type Option func(*options)

type options struct {
	bg color.Color // background color
}

// Background sets the background color of a layout.
func Background(bg color.Color) Option {
	return func(o *options) {
		o.bg = bg
	}
}

// EvalOptions evaluates a list of option functions, returning the final set.
func evalOptions(o ...Option) options {
	opts := options{
		bg: defaultBg,
	}
	for i := range o {
		o[i](&opts)
	}
	return opts
}
