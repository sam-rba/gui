package strain_test

import (
	"testing"
	"image"
	"slices"

	"github.com/lithdew/casso"

	"github.com/faiface/gui/style"
	"github.com/faiface/gui/lay/strain"
)

type solverTest struct {
	t *testing.T
	*style.Style
	*strain.Solver
}

func newSolverTest(t *testing.T, constraints []<-chan strain.Constraint) solverTest {
	styl, err := style.New()
	if err != nil {
		t.Fatal(err)
	}
	solver, err := strain.NewSolver(styl, constraints)
	if err != nil {
		t.Fatal(err)
	}
	return solverTest{t, styl, solver}
}

func (st solverTest) Close() {
	st.Solver.Close()
	if err := st.Style.Close(); err != nil {
		st.t.Error(err)
	}
}

func (st solverTest) addConstraint(op casso.Op, lhs, rhs casso.Symbol) {
	if err := st.Solver.AddConstraint(op, lhs, rhs); err != nil {
		st.t.Error(err)
	}
}

func (st solverTest) solve(container image.Rectangle, wantFields []image.Rectangle) {
	fields, err := st.Solver.Solve(container)
	if err != nil {
		st.t.Errorf("Solve(%v): %v; want %v", container, err, wantFields)
	}
	if !slices.Equal(fields, wantFields) {
		st.t.Errorf("Solve(%v) = %v; want %v", container, fields, wantFields)
	}
}

// No constraints and zero-sized container.
func TestTrivial(t *testing.T) {
	t.Parallel()
	st := newSolverTest(t, nil)
	defer st.Close()
	fields, err := st.Solver.Solve(image.ZR)
	if err != nil {
		t.Error(err)
	}
	if len(fields) != 0 {
		t.Errorf("expected 0 fields; got %d", len(fields))
	}
}

// One field that occupies the whole container.
func TestSingleField(t *testing.T) {
	// Setup
	t.Parallel()
	constraints := make(chan strain.Constraint)
	st := newSolverTest(t, []<-chan strain.Constraint{constraints})
	defer st.Close()
	defer close(constraints)

	// Add constraints
	container := st.Solver.Container()
	field := st.Solver.Field(0)
	st.addConstraint(casso.EQ, field.Origin.X, container.Origin.X)
	st.addConstraint(casso.EQ, field.Origin.Y, container.Origin.Y)
	st.addConstraint(casso.EQ, field.Size.X, container.Size.X)
	st.addConstraint(casso.EQ, field.Size.Y, container.Size.Y)

	// Solve
	for _, container := range []image.Rectangle{
		image.ZR,
		image.Rectangle{image.ZP, image.Pt(800, 600)},
		image.Rectangle{image.Pt(12, 34), image.Pt(123, 456)},
	} {
		// field == container
		st.solve(container, []image.Rectangle{container})
	}
}

// Solver with only layout constaints, no field constraints.
func TestLayConstrs(t *testing.T) {
	t.Parallel()

	st := newSolverTest(t, nil)
	defer st.Close()

	t.Fail() // TODO
}
