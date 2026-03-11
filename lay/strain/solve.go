package strain

import (
	"context"
	"fmt"
	"image"
	"math"
	"sync"

	"github.com/lithdew/casso"

	"github.com/faiface/gui/internal/log"
	"github.com/faiface/gui/internal/tag"
	"github.com/faiface/gui/style"
)

const (
	solverPriority = casso.Medium
	layoutPriority = casso.Medium
	fieldPriority  = casso.Medium
)

// solver wraps Solver and protects the not-thread-safe casso.Solver.
type solver struct {
	s                *Solver
	cs               *casso.Solver
	style            *style.Style
	fieldSizeConstrs []sizeConstraintSymbols
	ctx              context.Context
}

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
	// External symbols
	container SymRect   // position and size of container
	fields    []SymRect // position and size of each field

	fieldConstrs  chan tag.Tagged[Constraint, fieldIndex]
	layoutConstrs chan constrainRequest
	solveReqs     chan solveRequest
}

// sizeConstraintSymbols is a set of constraint ID symbols that
// control the upper/lower/exact bounds on the width and height of a
// field.
type sizeConstraintSymbols struct {
	widthEq, widthGte, widthLte    *casso.Symbol
	heightEq, heightGte, heightLte *casso.Symbol

	// These are the return values of casso.Solver.AddConstraint().
	// Eq is mutually exclusive with both gte and lte. But gte and lte are
	// not mutually exclusive (can have both a lower and upper bound).
	// Nil means no constraint.
}

type fieldIndex int

type constrainRequest struct {
	casso.Priority
	casso.Op
	constant float64
	terms    []casso.Term
	res      chan<- error
}

type solveRequest struct {
	container image.Rectangle
	res       chan<- solveResponse
}

type solveResponse struct {
	fields []image.Rectangle
	error
}

// NewSolver creates a Solver that can be used to resolve constraints
// received from the given channels: one channel per field in the
// layout. These are generally the receiving side of some Envs'
// Impose() channels.
func NewSolver(styl *style.Style, constraints []<-chan Constraint) (*Solver, error) {
	nfields := len(constraints)
	fields := make([]SymRect, nfields)
	fieldConstrs := make(chan tag.Tagged[Constraint, fieldIndex])
	var wg sync.WaitGroup
	wg.Add(nfields)
	ctx, cancel := context.WithCancel(context.Background())
	for i := range fields {
		fields[i] = NewSymRect()
		go func() {
			// Tag incoming field constraints by field index and multiplex them into fieldConstrs
			tag.Tag(ctx, fieldConstrs, constraints[i], func(c Constraint) fieldIndex {
				return fieldIndex(i)
			})
			wg.Done()
		}()
	}
	go func() {
		wg.Wait() // once all fields close their constraints channels,
		cancel()  // destroy the solver
	}()

	cs := casso.NewSolver()
	container := NewSymRect()
	if err := editRect(cs, container, casso.Strong); err != nil {
		cancel()
		return nil, fmt.Errorf("error marking container symbol as editable: %w", err)
	}

	s := &solver{
		&Solver{
			container,
			fields,
			fieldConstrs,
			make(chan constrainRequest),
			make(chan solveRequest),
		},
		cs,
		styl,
		make([]sizeConstraintSymbols, nfields),
		ctx,
	}
	go s.run()
	if err := s.s.addDefaultConstraints(); err != nil {
		cancel()
		return nil, fmt.Errorf("error adding default constraint: %w", err)
	}
	return s.s, nil
}

func (s *solver) run() {
	defer close(s.s.fieldConstrs)
	defer close(s.s.layoutConstrs)
	defer close(s.s.solveReqs)

	for {
		select {
		case tc := <-s.s.fieldConstrs:
			constr, i := tc.Val, tc.Tag
			if err := s.addFieldSizeConstraint(constr, s.s.fields[i], &s.fieldSizeConstrs[i]); err != nil {
				log.Err.Printf("error adding layout constraint %#v from field %d: %v\n",
					constr, i, err)
			}

		case req := <-s.s.layoutConstrs:
			_, err := s.addConstraint(req.Priority, req.Op, req.constant, req.terms...)
			req.res <- err

		case req := <-s.s.solveReqs:
			fields, err := s.solve(req.container)
			req.res <- solveResponse{fields, err}

		case <-s.ctx.Done():
			return
		}
	}
}

// addFieldSizeConstraint adds or modifies a constraint on the size of
// a field, removing mutually exclusive constraints.
func (s *solver) addFieldSizeConstraint(constr Constraint, field SymRect, fieldConstrs *sizeConstraintSymbols) error {
	// Clear mutually exclusive constraints and replace with new one
	switch constr.Dimension {
	case Width:
		width := float64(s.style.Pixels(constr.Value).Round())
		switch constr.Op {
		case casso.EQ:
			s.removeConstraints(fieldConstrs.widthEq, fieldConstrs.widthGte, fieldConstrs.widthLte)
			c, err := s.addConstraint(fieldPriority, constr.Op, -width, field.Size.X.T(1.0))
			if err != nil {
				return err
			}
			fieldConstrs.widthEq = &c
		case casso.GTE:
			s.removeConstraints(fieldConstrs.widthEq, fieldConstrs.widthGte)
			c, err := s.addConstraint(fieldPriority, constr.Op, -width, field.Size.X.T(1.0))
			if err != nil {
				return err
			}
			fieldConstrs.widthGte = &c
		case casso.LTE:
			s.removeConstraints(fieldConstrs.widthEq, fieldConstrs.widthLte)
			c, err := s.addConstraint(fieldPriority, constr.Op, -width, field.Size.X.T(1.0))
			if err != nil {
				return err
			}
			fieldConstrs.widthLte = &c
		default:
			panic(fmt.Sprintf("unreachable: impossible %T: %v", constr.Op, constr.Op))
		}
	case Height:
		height := float64(s.style.Pixels(constr.Value).Round())
		switch constr.Op {
		case casso.EQ:
			s.removeConstraints(fieldConstrs.heightEq, fieldConstrs.heightGte, fieldConstrs.heightLte)
			c, err := s.addConstraint(fieldPriority, constr.Op, -height, field.Size.Y.T(1.0))
			if err != nil {
				return err
			}
			fieldConstrs.heightEq = &c
		case casso.GTE:
			s.removeConstraints(fieldConstrs.heightEq, fieldConstrs.heightGte)
			c, err := s.addConstraint(fieldPriority, constr.Op, -height, field.Size.Y.T(1.0))
			if err != nil {
				return err
			}
			fieldConstrs.heightGte = &c
		case casso.LTE:
			s.removeConstraints(fieldConstrs.heightEq, fieldConstrs.heightLte)
			c, err := s.addConstraint(fieldPriority, constr.Op, -height, field.Size.Y.T(1.0))
			if err != nil {
				return err
			}
			fieldConstrs.heightLte = &c
		default:
			panic(fmt.Sprintf("unreachable: impossible %T: %v", constr.Op, constr.Op))
		}
	default:
		panic(fmt.Sprintf("unreachable: impossible %T: %v", constr.Dimension, constr.Dimension))
	}

	return nil
}

func (s *solver) removeConstraints(constrs ...*casso.Symbol) error {
	for _, constr := range constrs {
		if constr != nil {
			if err := s.cs.RemoveConstraint(*constr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *solver) addConstraint(priority casso.Priority, op casso.Op, constant float64, terms ...casso.Term) (casso.Symbol, error) {
	return s.cs.AddConstraintWithPriority(priority, casso.NewConstraint(op, constant, terms...))
}

func (s *solver) solve(container image.Rectangle) (fields []image.Rectangle, err error) {
	if err := suggestRect(s.cs, s.s.container, container); err != nil {
		return nil, err
	}

	fields = make([]image.Rectangle, len(s.s.fields))
	for i, field := range s.s.fields {
		min := image.Pt(
			s.val(field.Origin.X),
			s.val(field.Origin.Y))
		max := min.Add(image.Pt(
			s.val(field.Size.X),
			s.val(field.Size.Y)))
		fields[i] = image.Rectangle{min, max}
	}
	return fields, nil
}

func (s *solver) val(sym casso.Symbol) int {
	return int(math.Round(s.cs.Val(sym)))
}

func (s *Solver) addDefaultConstraints() error {
	for _, field := range s.fields {
		if err := s.addConstraintPt(solverPriority, casso.GTE, field.Origin, s.container.Origin); err != nil {
			return err
		}
		if err := s.addConstraintPt(solverPriority, casso.LTE, field.Size, s.container.Size); err != nil {
			return err
		}
	}
	return nil
}

// Container returns the Cassowary symbols representing the layout
// container's position and size.
func (s *Solver) Container() SymRect { return s.container }

// Field returns the Cassowary symbols representing the i'th field's
// position and size.
func (s *Solver) Field(i int) SymRect { return s.fields[i] }

// AddConstraint imposes a constraint on a set of terms:
//
//	constant + ∑terms op 0
//
// The terms can be obtained by calling the T(coeff) method of the
// symbols returned by Container() or Field(n). Think of terms with
// positive coefficients as being on the LHS and terms with negative
// coefficients as being on the RHS.
//
// Once added, a constraint cannot be removed or modified.
func (s *Solver) AddConstraint(op casso.Op, constant float64, terms ...casso.Term) error {
	return s.addConstraint(layoutPriority, op, constant, terms...)
}

// AddConstraintPt imposes a constraint between two point symbols:
// (lhs.X op rhs.X) and (lhs.Y op rhs.Y).
func (s *Solver) AddConstraintPt(op casso.Op, lhs, rhs SymPt) error {
	return s.addConstraintPt(layoutPriority, op, lhs, rhs)
}

func (s *Solver) addConstraint(priority casso.Priority, op casso.Op, constant float64, terms ...casso.Term) error {
	resc := make(chan error)
	defer close(resc)
	s.layoutConstrs <- constrainRequest{priority, op, constant, terms, resc}
	return <-resc
}

func (s *Solver) addConstraintPt(priority casso.Priority, op casso.Op, lhs, rhs SymPt) error {
	if err := s.addConstraint(priority, op, 0, lhs.X.T(1.0), rhs.X.T(-1.0)); err != nil {
		return err
	}
	if err := s.addConstraint(priority, op, 0, lhs.Y.T(1.0), rhs.Y.T(-1.0)); err != nil {
		return err
	}
	return nil
}

func (s *Solver) addSizeConstraint(priority casso.Priority, op casso.Op, lhs SymPt, rhs image.Point) error {
	if err := s.addConstraint(priority, op, -float64(rhs.X), lhs.X.T(1.0)); err != nil {
		return err
	}
	if err := s.addConstraint(priority, op, -float64(rhs.Y), lhs.Y.T(1.0)); err != nil {
		return err
	}
	return nil
}

// Solve uses the constraints that the Solver has received so far to
// partition container, dividing it among the fields of the layout.
//
// It returns a slice of Rectangles, one per field. The slice has the
// same length as the slice of Constraint channels that were passed to
// NewSolver().
func (s *Solver) Solve(container image.Rectangle) (fields []image.Rectangle, err error) {
	resc := make(chan solveResponse)
	defer close(resc)
	s.solveReqs <- solveRequest{container, resc}
	res := <-resc
	return res.fields, res.error
}
