package strain

import (
	"context"
	"fmt"
	"image"

	"github.com/lithdew/casso"

	"github.com/faiface/gui/internal/log"
	"github.com/faiface/gui/internal/tag"
	"github.com/faiface/gui/style"
)

// Solver uses the Cassowary algorithm to partition a rectangle among
// several layout fields.
//
// Solver uses constraints to control the position and size of fields.
// There are two sources of constraints: 1) the layout, via
// AddConstraint(); and 2) the fields, via the Constraint channels
// passed to NewSolver(). The layout constraints control the position
// and size of fields within the container, while the field
// constraints control the width and height of each field.
type Solver struct {
	solver *casso.Solver

	// External symbols
	container    SymRect // position and size of container
	fieldOrigins []SymPt // top-left corner position of each field
	fieldSizes   []SymPt // width and height of each field

	fieldSizeConstrs []sizeConstraint

	fieldConstrs  chan tag.Tagged[Constraint, int] // constraints from fields, tagged by index
	layoutConstrs chan constrainRequest            // constraints from the layout via AddConstraint()
	solveReqs     chan solveRequest

	style style.Style

	ctx    context.Context
	cancel func()
}

// sizeConstraint sets upper/lower/exact bounds on the width and
// height of a field.
type sizeConstraint struct {
	widthEq, widthGte, widthLte    *casso.Symbol
	heightEq, heightGte, heightLte *casso.Symbol

	// These are the return values of casso.Solver.AddConstraint().
	// Eq is mutually exclusive with both gte and lte. But gte and lte are
	// not mutually exclusive (can have both a lower and upper bound).
	// Nil means no constraint.
}

type constrainRequest struct {
	constr casso.Constraint
	res    chan<- error
}

type solveRequest struct {
	container image.Rectangle
	res       chan<- solveResponse
}

type solveResponse struct {
	fields []image.Rectangle
	err    error
}

// NewSolver creates a Solver that can be used to resolve constraints
// received from the given channels; one channel per field in the
// layout. These are generally the receiving side of some Envs'
// Impose() channels.
//
// The Solver should be closed after use, but not before all of the
// constraint channels have been closed.
func NewSolver(styl style.Style, constraints []<-chan Constraint) (*Solver, error) {
	nfields := len(constraints)

	fieldConstrs := make(chan tag.Tagged[Constraint, int])
	fieldOrigins := make([]SymPt, nfields)
	fieldSizes := make([]SymPt, nfields)
	for i, cs := range constraints {
		go tag.Tag(fieldConstrs, cs, func(c Constraint) int { return i })
		fieldOrigins[i] = NewSymPt()
		fieldSizes[i] = NewSymPt()
	}

	ctx, cancel := context.WithCancel(context.Background())
	solver := &Solver{
		casso.NewSolver(),
		NewSymRect(),
		fieldOrigins,
		fieldSizes,
		make([]sizeConstraint, nfields),
		fieldConstrs,
		make(chan constrainRequest),
		make(chan solveRequest),
		styl,
		ctx,
		cancel,
	}
	if err := editRect(solver.solver, solver.container, casso.Required); err != nil {
		return nil, err
	}

	go solver.run()

	return solver, nil
}

func (s *Solver) run() {
	defer close(s.fieldConstrs)
	defer close(s.layoutConstrs)
	defer close(s.solveReqs)

	for {
		select {
		case tc := <-s.fieldConstrs: // constraint from a field, tagged by field index
			constr, fieldIdx := tc.Val, tc.Tag
			err := s.addSizeConstraint(constr, fieldIdx)
			if err != nil {
				log.Err.Printf("error adding layout constraint %#v from field %d: %v\n",
					constr, fieldIdx, err)
			}

		case req := <-s.layoutConstrs: // constraint from the layout via AddConstraint()
			_, err := s.solver.AddConstraint(req.constr)
			req.res <- err

		case req := <-s.solveReqs:
			fields, err := s.solve(req.container)
			req.res <- solveResponse{fields, err}

		case <-s.ctx.Done():
			return
		}
	}
}

// addSizeConstraint adds or modifies a constraint on the size of a
// field, removing mutually exclusive constraints.
func (s *Solver) addSizeConstraint(constr Constraint, fieldIdx int) error {
	fieldSize := s.fieldSizes[fieldIdx]
	fieldConstrs := &s.fieldSizeConstrs[fieldIdx]

	// Clear mutually exclusive constraints and replace with new one
	switch constr.Dim {
	case Width:
		width := s.style.Pixels(constr.Value).Round()
		switch constr.Op {
		case casso.EQ:
			s.removeConstraints(fieldConstrs.widthEq, fieldConstrs.widthGte, fieldConstrs.widthLte)
			c, err := s.solver.AddConstraint(casso.NewConstraint(constr.Op, -width, fieldSize.X.T(1.0)))
			if err != nil {
				return err
			}
			fieldConstrs.widthEq = &c
		case casso.GTE:
			s.removeConstraints(fieldConstrs.widthEq, fieldConstrs.widthGte)
			c, err := s.solver.AddConstraint(casso.NewConstraint(constr.Op, -width, fieldSize.X.T(1.0)))
			if err != nil {
				return err
			}
			fieldConstrs.widthGte = &c
		case casso.LTE:
			s.removeConstraints(fieldConstrs.widthEq, fieldConstrs.widthLte)
			c, err := s.solver.AddConstraint(casso.NewConstraint(constr.Op, -width, fieldSize.X.T(1.0)))
			if err != nil {
				return err
			}
			fieldConstrs.widthLte = &c
		default:
			panic(fmt.Sprintf("unreachable: impossible %T: %v", constr.Op, constr.Op))
		}
	case Height:
		height := s.style.Pixels(constr.Value).Round()
		switch constr.Op {
		case casso.EQ:
			s.removeConstraints(fieldConstrs.heightEq, fieldConstrs.heightGte, fieldConstrs.heightLte)
			c, err := s.solver.AddConstraint(casso.NewConstraint(constr.Op, -height, fieldSize.Y.T(1.0)))
			if err != nil {
				return err
			}
			fieldConstrs.heightEq = &c
		case casso.GTE:
			s.removeConstraints(fieldConstrs.heightEq, fieldConstrs.heightGte)
			c, err := s.solver.AddConstraint(casso.NewConstraint(constr.Op, -height, fieldSize.Y.T(1.0)))
			if err != nil {
				return err
			}
			fieldConstrs.heightGte = &c
		case casso.LTE:
			s.removeConstraints(fieldConstrs.heightEq, fieldConstrs.heightLte)
			c, err := s.solver.AddConstraint(casso.NewConstraint(constr.Op, -height, fieldSize.Y.T(1.0)))
			if err != nil {
				return err
			}
			fieldConstrs.heightLte = &c
		default:
			panic(fmt.Sprintf("unreachable: impossible %T: %v", constr.Op, constr.Op))
		}
	default:
		panic(fmt.Sprintf("unreachable: impossible %T: %v", constr.Dim, constr.Dim))
	}

	return nil
}

func (s *Solver) removeConstraints(constrs ...*casso.Symbol) error {
	for _, constr := range constrs {
		if constr != nil {
			if err := s.solver.RemoveConstraint(*constr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Solver) solve(container image.Rectangle) (fields []image.Rectangle, err error) {
	if err := suggestRect(s.solver, s.container, container); err != nil {
		return nil, err
	}
	fields = make([]image.Rectangle, len(s.fieldOrigins))
	for i := range fields {
		origin, size := s.fieldOrigins[i], s.fieldSizes[i]
		min := image.Pt(int(s.solver.Val(origin.X)), int(s.solver.Val(origin.Y)))
		max := min.Add(image.Pt(int(s.solver.Val(size.X)), int(s.solver.Val(size.Y))))
		fields[i] = image.Rectangle{min, max}
	}
	return fields, nil
}

// Close destroys the solver. You must first close all the Constraint
// channels that were passed to NewSolver() before calling Close.
func (s *Solver) Close() { s.cancel() }

// Container returns the Cassowary symbols representing the layout
// container's position and size.
func (s *Solver) Container() SymRect { return s.container }

// Field returns the Cassowary symbols representing the i'th field's
// position and size. origin represents the position of the top-left
// corner of the field. size represents the width and height of the
// field.
func (s *Solver) Field(i int) (origin, size SymPt) { return s.fieldOrigins[i], s.fieldSizes[i] }

// AddConstraint imposes a constraint between two symbols. The symbols
// may be aspects of the Container() or Field()s.
//
// Once added, a constraint cannot be removed or modified.
func (s *Solver) AddConstraint(op casso.Op, lhs, rhs casso.Symbol) error {
	res := make(chan error)
	defer close(res)
	s.layoutConstrs <- constrainRequest{
		casso.NewConstraint(op, 0, lhs.T(1.0), rhs.T(-1.0)),
		res,
	}
	return <-res
}

// Solve uses the constraints that the Solver has received so far to
// partition container, dividing it among the fields of the layout.
//
// It returns a slice of Rectangles, one per field. The slice has the
// same length as the slice of Constraint channels that were passed to
// NewSolver().
func (s *Solver) Solve(container image.Rectangle) ([]image.Rectangle, error) {
	resc := make(chan solveResponse)
	defer close(resc)
	s.solveReqs <- solveRequest{container, resc}
	res := <-resc
	return res.fields, res.err
}
