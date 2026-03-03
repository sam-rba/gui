package strain

import (
	"image"

	"github.com/lithdew/casso"
)

// SymPt is a set of Cassowary symbols representing an image.Point.
type SymPt struct {
	X, Y casso.Symbol
}

// SymRect is a set of Cassowary symbols representing an
// image.Rectangle.
type SymRect struct {
	Origin SymPt // top-left corner position
	Size   SymPt // Dx() and Dy()
}

func NewSymPt() SymPt {
	return SymPt{casso.New(), casso.New()}
}

func NewSymRect() SymRect {
	return SymRect{NewSymPt(), NewSymPt()}
}

// editRect marks all symbols of a rectangle as editable with a
// certain precedence.
func editRect(solver *casso.Solver, sr SymRect, p casso.Priority) error {
	if err := editPt(solver, sr.Origin, p); err != nil {
		return err
	}
	if err := editPt(solver, sr.Size, p); err != nil {
		return err
	}
	return nil
}

// editRect marks both symbols of a point as editable with a certain
// precedence.
func editPt(solver *casso.Solver, sp SymPt, p casso.Priority) error {
	if err := solver.Edit(sp.X, p); err != nil {
		return err
	}
	if err := solver.Edit(sp.Y, p); err != nil {
		return err
	}
	return nil
}

func suggestRect(solver *casso.Solver, sr SymRect, r image.Rectangle) error {
	if err := suggestPt(solver, sr.Origin, r.Min); err != nil {
		return err
	}
	if err := suggestPt(solver, sr.Size, image.Pt(r.Dx(), r.Dy())); err != nil {
		return err
	}
	return nil
}

func suggestPt(solver *casso.Solver, sp SymPt, p image.Point) error {
	if err := solver.Suggest(sp.X, float64(p.X)); err != nil {
		return err
	}
	if err := solver.Suggest(sp.Y, float64(p.Y)); err != nil {
		return err
	}
	return nil
}
