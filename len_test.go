package gui_test

import (
	"fmt"
	"image"

	"github.com/faiface/gui"
)

func ExampleRelative() {
	var l Length = Relative(0.10)   // 10%
	r := image.Rect(0, 0, 100, 100) // 100x100 rectangle
	fmt.Println(l.Px(r))
	// Output:
	// 10
}
