package layout_test

import (
	"image"
	"testing"

	"github.com/faiface/gui/layout"
)

func TestResizeAll(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	child := layout.ResizeAll(parent)
	want := parent
	if child != want {
		t.Errorf("got %v; want %v", child, want)
	}
}

func TestResizeQuad1(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	child := layout.ResizeQuad1(parent)
	want := image.Rect(222, 222, 333, 333)
	if child != want {
		t.Errorf("got %v; want %v", child, want)
	}
}

func TestResizeQuad2(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	child := layout.ResizeQuad2(parent)
	want := image.Rect(111, 222, 222, 333)
	if child != want {
		t.Errorf("got %v; want %v", child, want)
	}
}

func TestResizeQuad3(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	child := layout.ResizeQuad3(parent)
	want := image.Rect(111, 333, 222, 444)
	if child != want {
		t.Errorf("got %v; want %v", child, want)
	}
}

func TestResizeQuad4(t *testing.T) {
	t.Parallel()
	parent := image.Rect(111, 222, 333, 444)
	child := layout.ResizeQuad4(parent)
	want := image.Rect(222, 333, 333, 444)
	if child != want {
		t.Errorf("got %v; want %v", child, want)
	}
}
