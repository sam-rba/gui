package strain

import (
	"github.com/lithdew/casso"
	"golang.org/x/exp/shiny/unit"
)

// Constraint imposes a restriction on the size of a widget or layout.
type Constraint struct {
	// Dimension is the dimension to constrain: width/height.
	Dimension

	// Relation declares whether the constraint is an upper, lower,
	// or exact bound.
	casso.Op

	// Value is the target or threshold value.
	unit.Value
}

// Dim is a dimension of a widget or layout that can be constrained.
type Dimension int

const (
	_ Dimension = iota
	Width
	Height
)
