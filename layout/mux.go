package layout

import (
	"image"
	"image/draw"
)

// TaggedDrawCall is a draw function tagged with the index of a child Env within a layout.
type taggedDrawCall struct {
	f   func(draw.Image) image.Rectangle
	idx uint
}

// MuxDrawCalls tags draw functions with the index of a child Env, and forwards them to an output channel.
// When the input channel is closed, it decrements the waitgroup counter.
// It does NOT close the output channel.
func muxDrawCalls(in <-chan func(draw.Image) image.Rectangle, idx uint, out chan<- taggedDrawCall, wg waitgroup) {
	for f := range in {
		out <- taggedDrawCall{f, idx}
	}
	wg.Done()
}

// Multicast sends a message to multiple recipients.
func multicast[T any](msg T, recipients []chan T) {
	for _, c := range recipients {
		c <- msg
	}
}
