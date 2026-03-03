package strain

import (
	"fmt"
	"image"
	"sync"

	"github.com/lithdew/casso"

	"github.com/faiface/gui/internal/log"
	"github.com/faiface/gui/internal/tag"
	"github.com/faiface/gui/style"
)

const (
	fieldConstraintPriority  = casso.Medium
	layoutConstraintPriority = casso.Medium
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
	solver       *casso.Solver
	style        *style.Style
	fieldConstrs chan tag.Tagged[Constraint, fieldIndex] // incoming constraints from fields

	// External symbols
	container SymRect   // position and size of container
	fields    []SymRect // position and size of each field

	// Constraint ID symbols
	fieldSizeConstrs []sizeConstraintSymbols

	mu sync.Mutex
}

type fieldIndex int

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
	for i, cs := range constraints {
		fields[i] = NewSymRect()
		go func() {
			// Tag incoming field constraints by field index and multiplex them into fieldConstrs
			tag.Tag(fieldConstrs, cs, func(c Constraint) fieldIndex {
				return fieldIndex(i)
			})
			wg.Done()
		}()
	}

	go func() {
		wg.Wait() // once all fields close their constraints channels,
		close(fieldConstrs)
	}()

	solver := casso.NewSolver()
	container := NewSymRect()
	if err := editRect(solver, container, casso.Strong); err != nil {
		return nil, fmt.Errorf("error marking container symbol as editable: %w", err)
	}

	s := &Solver{
		solver,
		styl,
		fieldConstrs,
		container,
		fields,
		make([]sizeConstraintSymbols, nfields),
		sync.Mutex{},
	}

	go func() {
		for tc := range fieldConstrs {
			constr, fieldIdx := tc.Val, tc.Tag
			s.mu.Lock()
			err := s.addFieldSizeConstraint(constr, fieldIdx)
			s.mu.Unlock()
			if err != nil {
				log.Err.Printf("error adding layout constraint %#v from field %d: %v\n",
					constr, fieldIdx, err)
			}
		}
	}()

	if err := s.addDefaultConstraints(); err != nil {
		return nil, fmt.Errorf("error adding default constraint: %w", err)
	}

	return s, nil
}

func (s *Solver) addDefaultConstraints() error {
	for _, field := range s.fields {
		if err := s.AddConstraintPt(casso.GTE, field.Origin, s.container.Origin); err != nil {
			return err
		}
		if err := s.AddConstraintPt(casso.LTE, field.Size, s.container.Size); err != nil {
			return err
		}
	}
	return nil
}

// addSizeConstraint adds or modifies a constraint on the size of a
// field, removing mutually exclusive constraints.
//
// Solver.mu must be held.
func (s *Solver) addFieldSizeConstraint(constr Constraint, i fieldIndex) error {
	fieldSize := s.fields[i].Size
	fieldConstrs := &s.fieldSizeConstrs[i]

	// Clear mutually exclusive constraints and replace with new one
	switch constr.Dimension {
	case Width:
		width := float64(s.style.Pixels(constr.Value).Round())
		switch constr.Op {
		case casso.EQ:
			s.removeConstraints(fieldConstrs.widthEq, fieldConstrs.widthGte, fieldConstrs.widthLte)
			c, err := s.addFieldConstraint(constr.Op, -width, fieldSize.X.T(1.0))
			if err != nil {
				return err
			}
			fieldConstrs.widthEq = &c
		case casso.GTE:
			s.removeConstraints(fieldConstrs.widthEq, fieldConstrs.widthGte)
			c, err := s.addFieldConstraint(constr.Op, -width, fieldSize.X.T(1.0))
			if err != nil {
				return err
			}
			fieldConstrs.widthGte = &c
		case casso.LTE:
			s.removeConstraints(fieldConstrs.widthEq, fieldConstrs.widthLte)
			c, err := s.addFieldConstraint(constr.Op, -width, fieldSize.X.T(1.0))
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
			c, err := s.addFieldConstraint(constr.Op, -height, fieldSize.Y.T(1.0))
			if err != nil {
				return err
			}
			fieldConstrs.heightEq = &c
		case casso.GTE:
			s.removeConstraints(fieldConstrs.heightEq, fieldConstrs.heightGte)
			c, err := s.addFieldConstraint(constr.Op, -height, fieldSize.Y.T(1.0))
			if err != nil {
				return err
			}
			fieldConstrs.heightGte = &c
		case casso.LTE:
			s.removeConstraints(fieldConstrs.heightEq, fieldConstrs.heightLte)
			c, err := s.addFieldConstraint(constr.Op, -height, fieldSize.Y.T(1.0))
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

// Solver.mu must be held.
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

// Solver.mu must be held.
func (s *Solver) addFieldConstraint(op casso.Op, constant float64, terms ...casso.Term) (casso.Symbol, error) {
	return s.solver.AddConstraintWithPriority(
		fieldConstraintPriority,
		casso.NewConstraint(op, constant, terms...))
}

// Container returns the Cassowary symbols representing the layout
// container's position and size.
func (s *Solver) Container() SymRect { return s.container }

// Field returns the Cassowary symbols representing the i'th field's
// position and size.
func (s *Solver) Field(i int) SymRect { return s.fields[i] }

// AddConstraint imposes a constraint between two symbols. The symbols
// may be aspects of the Container() or Field()s.
//
// Once added, a constraint cannot be removed or modified.
func (s *Solver) AddConstraint(op casso.Op, lhs, rhs casso.Symbol) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.solver.AddConstraintWithPriority(layoutConstraintPriority, casso.NewConstraint(op, 0, lhs.T(1.0), rhs.T(-1.0)))
	return err
}

// AddConstraintPt imposes a constraint between two point symbols:
// (lhs.X op rhs.X) and (lhs.Y op rhs.Y).
func (s *Solver) AddConstraintPt(op casso.Op, lhs, rhs SymPt) error {
	if err := s.AddConstraint(op, lhs.X, rhs.X); err != nil {
		return err
	}
	if err := s.AddConstraint(op, lhs.Y, rhs.Y); err != nil {
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
func (s *Solver) Solve(container image.Rectangle) ([]image.Rectangle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := suggestRect(s.solver, s.container, container); err != nil {
		return nil, err
	}

	fields := make([]image.Rectangle, len(s.fields))
	for i := range fields {
		field := s.fields[i]
		min := image.Pt(
			s.intVal(field.Origin.X),
			s.intVal(field.Origin.Y))
		max := min.Add(image.Pt(
			s.intVal(field.Size.X),
			s.intVal(field.Size.Y)))
		fields[i] = image.Rectangle{min, max}
	}
	return fields, nil
}

// Solver.mu must be held.
func (s *Solver) intVal(sym casso.Symbol) int { return int(s.solver.Val(sym)) }
