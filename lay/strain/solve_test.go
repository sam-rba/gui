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
	defer close(constraints)
	defer st.Close()

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
		defer close(constraints)
		defer st.Close()

		// Add field constraints
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
		defer close(constraints)
		defer st.Close()

		// Add field constraints
		constraints <- strain.Constraint{strain.Width, casso.GTE, unit.Value{200, unit.Px}}
		constraints <- strain.Constraint{strain.Height, casso.GTE, unit.Value{300, unit.Px}}

		// Solve
		container := image.Rect(12, 34, 100, 200)
		synctest.Wait()
		st.solve(container, validateEq([]image.Rectangle{container}))
	})
}

// Fields arranged as rows.
func TestRows(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		// Setup
		nrows := 8
		constraintss := make([]chan strain.Constraint, nrows)
		for i := range constraintss {
			constraintss[i] = make(chan strain.Constraint)
		}
		st := newSolverTest(t, castRx(constraintss))
		defer func() {
			for _, c := range constraintss {
				close(c)
			}
			st.Close()
		}()

		// Add layout constraints
		require.NoError(t, st.Solver.AddConstraintPt(casso.EQ, st.Solver.Field(0).Origin, st.Solver.Container().Origin)) // start from top left corner
		fieldHeights := make([]casso.Term, nrows)
		for i := 0; i < nrows; i++ {
			fieldHeights[i] = st.Field(i).Size.Y.T(1)
			container := st.Solver.Container()
			f := st.Solver.Field(i)
			require.NoError(t, st.Solver.AddConstraint(casso.EQ, 0, f.Size.X.T(1), container.Size.X.T(-1))) // span full width
		}
		terms := append(fieldHeights, st.Solver.Container().Size.Y.T(-1)) // ∑field[i].height <= container.height
		require.NoError(t, st.Solver.AddConstraint(casso.LTE, 0, terms...))
		for i := 1; i < nrows; i++ {
			f0, f1 := st.Solver.Field(i-1), st.Solver.Field(i)
			require.NoError(t, st.Solver.AddConstraint(casso.EQ, 0, f1.Origin.Y.T(1), f0.Origin.Y.T(-1), f0.Size.Y.T(-1))) // in order
			require.NoError(t, st.Solver.AddConstraintPt(casso.EQ, f1.Size, f0.Size))                                      // same size
		}

		// Add field constraints
		var rowHeight int
		for i, c := range constraintss {
			rowHeight = 2 * i
			c <- strain.Constraint{strain.Width, casso.GTE, unit.Value{float64(16 + i), unit.Ch}}
			c <- strain.Constraint{strain.Height, casso.GTE, unit.Value{float64(rowHeight), unit.Px}}
		}

		// Solve
		container := image.Rect(123, 234, 567, 678)
		synctest.Wait()
		fields, err := st.Solver.Solve(container)
		if err != nil {
			t.Fatal(err)
		}
		require.EqualValues(t, nrows, len(fields), "wrong number of fields")
		for _, field := range fields {
			t.Logf("%v (%d %d)\n", field, field.Dx(), field.Dy())
		}
		for i, field := range fields {
			require.Equal(t, container.Min.X, field.Min.X, "not left-aligned with container")
			require.Equal(t, container.Min.Y+i*rowHeight, field.Min.Y, "wrong y position")
			require.Equal(t, container.Dx(), field.Dx(), "wrong width")
			require.Equal(t, rowHeight, field.Dy(), "wrong height")
		}
	})
}

func TestSolver(t *testing.T) {
	t.Fail() // TODO: more tests
}

func castRx[T any](cs []chan T) []<-chan T {
	rcs := make([]<-chan T, len(cs))
	for i, c := range cs {
		rcs[i] = c
	}
	return rcs
}
