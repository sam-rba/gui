package gui_test

import (
	"fmt"

	"github.com/faiface/gui"
)

func ExampleRelative() {
	var l gui.Length = gui.Relative(0.10) // 10%
	fmt.Println(l.Px(100))
	// Output:
	// 10
}
