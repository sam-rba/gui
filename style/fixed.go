package style

import "golang.org/x/image/math/fixed"

func fixed26ToFloat(n fixed.Int26_6) float64 {
	return float64(n) / (1 << 6)
}

func floatToFixed26(f float64) fixed.Int26_6 {
	return fixed.Int26_6(f * (1 << 6))
}
