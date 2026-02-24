package gui

// TODO: add font-size-relative units once a text rendering package is added.

// Length allows distance or size to be expressed in absolute or relative units.
type Length interface {
	// Px resolves the Length to pixels for the given parent Env's size in pixels.
	Px(parent int) int
}

// Px is a Length expressed in pixels.
type Px int

// Px implements the Length interface.
func (p Px) Px(parent int) int { return int(p) }

// Relative is a Length relative to the width or height of the parent Env.
//
// Relative(0.10) <=> 10%.
type Relative float64

// Px implements the Length interface.
func (r Relative) Px(parent int) int {
	return int(float64(r) * float64(parent))
}
