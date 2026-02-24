package gui

// Constraint imposes a restriction on the size of a widget or layout.
type Constraint struct {
	// Dim is the dimension to constrain: width/height.
	Dim

	// Relation declares whether the constraint is an upper, lower, or exact bound.
	Relation

	// Length is the target or threshold value.
	Length
}

// Dim is a dimension of a widget or layout that can be constrained.
type Dim int

const (
	_ Dim = iota
	Width
	Height
)

// Relation is an (in)equality.
type Relation int

const (
	_    Relation = iota
	Eq            // ==
	Gteq          // >=
	Gt            // >
	Lteq          // <=
	Lt            // <
)
