package strain_test

import (
	"fmt"
	"image"
	"slices"
	"testing"
	"testing/synctest"

	"github.com/lithdew/casso"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/image/math/fixed"

	"github.com/faiface/gui/lay/strain"
	"github.com/faiface/gui/style"
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
	if err := st.Style.Close(); err != nil {
		st.t.Error(err)
	}
}

func (st solverTest) solve(container image.Rectangle, validate func(fields []image.Rectangle) error) {
	fields, err := st.Solver.Solve(container)
	if err != nil {
		st.t.Errorf("Solve(%v): %v", container, err)
	}
	if err := validate(fields); err != nil {
		st.t.Errorf("Solve(%v) = %v; %v", container, fields, err)
	}
}

func validateEq(wantFields []image.Rectangle) func(fields []image.Rectangle) error {
	return func(fields []image.Rectangle) error {
		if !slices.Equal(fields, wantFields) {
			return fmt.Errorf("want %v", wantFields)
		}
		return nil
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
	t.Parallel()

	// Setup
	constraints := make(chan strain.Constraint)
	st := newSolverTest(t, []<-chan strain.Constraint{constraints})
	defer st.Close()
	defer close(constraints)

	// Add layout constraints
	container := st.Solver.Container()
	field := st.Solver.Field(0)
	require.NoError(t, st.Solver.AddConstraintPt(casso.EQ, field.Origin, container.Origin))
	require.NoError(t, st.Solver.AddConstraintPt(casso.EQ, field.Size, container.Size))

	// Solve
	for _, container := range []image.Rectangle{
		image.ZR,
		image.Rectangle{image.ZP, image.Pt(800, 600)},
		image.Rectangle{image.Pt(12, 34), image.Pt(123, 456)},
	} {
		// field == container
		st.solve(container, validateEq([]image.Rectangle{container}))
	}
}

// Field gives its minimum size.
func TestFieldMinSize(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		// Setup
		constraints := make(chan strain.Constraint)
		st := newSolverTest(t, []<-chan strain.Constraint{constraints})
		defer st.Close()
		defer close(constraints)

		// Add widget constraints
		minWidth := unit.Value{32, unit.Ch}
		minHeight := unit.Value{1.5, unit.Em}
		constraints <- strain.Constraint{strain.Width, casso.GTE, minWidth}
		constraints <- strain.Constraint{strain.Height, casso.GTE, minHeight}
		synctest.Wait()

		// Solve
		st.solve(image.Rect(12, 34, 800, 600), func(fields []image.Rectangle) error {
			if len(fields) != 1 {
				return fmt.Errorf("got %d fields; want %d", len(fields), 1)
			}
			field := fields[0]
			if fixed.I(field.Dx()) < st.Style.Pixels(minWidth) {
				return fmt.Errorf("dx = %v; want >= %v", field.Dx(), st.Style.Pixels(minWidth))
			} else if fixed.I(field.Dy()) < st.Style.Pixels(minHeight) {
				return fmt.Errorf("dy = %v; want >= %v", field.Dy(), st.Style.Pixels(minHeight))
			}
			return nil
		})
	})
}

// Field min size larger than container.
func TestFieldMinSizeLargerThanContainer(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		// Setup
		constraints := make(chan strain.Constraint)
		st := newSolverTest(t, []<-chan strain.Constraint{constraints})
		defer st.Close()
		defer close(constraints)

		// Add widget constraints
		constraints <- strain.Constraint{strain.Width, casso.GTE, unit.Value{200, unit.Px}}
		constraints <- strain.Constraint{strain.Height, casso.GTE, unit.Value{300, unit.Px}}
		synctest.Wait()

		// Solve
		container := image.Rect(12, 34, 100, 200)
		st.solve(container, validateEq([]image.Rectangle{container}))
	})
}

// Solver with only layout constaints, no field constraints.
func TestLayConstrs(t *testing.T) {
	t.Parallel()

	st := newSolverTest(t, nil)
	defer st.Close()

	t.Fail() // TODO: more tests
}
