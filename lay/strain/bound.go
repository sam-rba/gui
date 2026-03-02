package strain

import "golang.org/x/exp/shiny/unit"

// Bounds places the inclusive lower and upper bounds [Min, Max]
// on a dimension.  A negative value means there is no bound.
// If both Min and Max are non-negative, then it is well formed if
// Min <= Max.
type Bounds struct {
	Min, Max unit.Value
}

// BoundedPoint is a point with bounds on its horizontal/vertical
// position.
type BoundedPoint struct {
	X, Y Bounds
}

// Free returns an unconstrained value (negative Min and Max).
func Free() Bounds {
	return Bounds{unit.Pixels(-1), unit.Pixels(-1)}
}

// FreePoint returns an unconstrained point (negative Min and Max
// in both dimensions).
func FreePoint() BoundedPoint {
	return BoundedPoint{Free(), Free()}
}
