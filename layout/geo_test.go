package layout_test

import (
	"image"
	"testing"

	"github.com/faiface/gui/layout"
)

func TestFull(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	want := parent
	testResize(t, layout.Full(), parent, want)
}

func TestQuad1(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	want := image.Rect(222, 222, 333, 333)
	testResize(t, layout.Quad1(), parent, want)
}

func TestQuad2(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	want := image.Rect(111, 222, 222, 333)
	testResize(t, layout.Quad2(), parent, want)
}

func TestQuad3(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	want := image.Rect(111, 333, 222, 444)
	testResize(t, layout.Quad3(), parent, want)
}

func TestQuad4(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	want := image.Rect(222, 333, 333, 444)
	testResize(t, layout.Quad4(), parent, want)
}

func TestInset(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	resize := layout.Inset(1, 2, 3, 4)
	want := image.Rect(115, 223, 331, 441)
	testResize(t, resize, parent, want)
}

func TestInsetAll(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	resize := layout.InsetAll(5)
	want := parent.Inset(5)
	testResize(t, resize, parent, want)
}

func TestInsetTop(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	resize := layout.InsetTop(2)
	want := image.Rect(111, 224, 333, 444)
	testResize(t, resize, parent, want)
}

func TestInsetRight(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	resize := layout.InsetRight(3)
	want := image.Rect(111, 222, 330, 444)
	testResize(t, resize, parent, want)
}

func TestInsetBottom(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	resize := layout.InsetBottom(4)
	want := image.Rect(111, 222, 333, 440)
	testResize(t, resize, parent, want)
}

func TestInsetLeft(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	resize := layout.InsetLeft(1)
	want := image.Rect(112, 222, 333, 444)
	testResize(t, resize, parent, want)
}

func testResize(t *testing.T, f layout.ResizeFunc, r, want image.Rectangle) {
	r1 := f(r)
	if r1 != want {
		t.Errorf("resize(%v) = %v; want %v", r, r1, want)
	}
}
