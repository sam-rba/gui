package style

import (
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/image/math/fixed"
)

// TODO
type Style struct {
}

// Convert implements the golang.org/x/exp/shiny/unit.Converter
// interface.
func (s *Style) Convert(v unit.Value, to unit.Unit) unit.Value {
	// TODO
}

// Pixels implements the golang.org/x/exp/shiny/unit.Converter
// interface.
func (s *Style) Pixels(v unit.Value) fixed.Int26_6 {
	// TODO
}
