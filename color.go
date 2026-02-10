package gui

import (
	"fmt"
	"image/color"
)

func HexToColor(hex string) color.Color {
	var r, g, b uint8
	fmt.Sscanf(hex, "#%2X%2X%2X", &r, &g, &b)
	return color.RGBA{r, g, b, 255}
}
