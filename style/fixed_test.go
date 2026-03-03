package style

import (
	"testing"

	"golang.org/x/image/math/fixed"
)

func TestFixed26ToFloat(t *testing.T) {
	t.Parallel()

	var (
		in        fixed.Int26_6
		out, want float64
	)

	in = 0
	out = fixed26ToFloat(in)
	want = 0
	if out != want {
		t.Errorf("float(%#v) = %v; want %v", in, out, want)
	}

	in = fixed.I(123)
	out = fixed26ToFloat(in)
	want = 123
	if out != want {
		t.Errorf("float(%#v) = %v; want %v", in, out, want)
	}

	in = fixed.I(-123)
	out = fixed26ToFloat(in)
	want = -123
	if out != want {
		t.Errorf("float(%#v) = %v; want %v", in, out, want)
	}

	in = fixed.Int26_6(1<<6 + 1<<4)
	out = fixed26ToFloat(in)
	want = 1.25
	if out != want {
		t.Errorf("float(%#v) = %v; want %v", in, out, want)
	}
}

func TestFloatToFixed26(t *testing.T) {
	t.Parallel()

	var (
		in        float64
		out, want fixed.Int26_6
	)

	in = 0
	out = floatToFixed26(in)
	want = 0
	if out != want {
		t.Errorf("fixed(%#v) = %v; want %v", in, out, want)
	}

	in = 1.25
	out = floatToFixed26(in)
	want = fixed.Int26_6(1<<6 + 1<<4)
	if out != want {
		t.Errorf("fixed(%#v) = %v; want %v", in, out, want)
	}

	in = -1.25
	out = floatToFixed26(in)
	want = -fixed.Int26_6(1<<6 + 1<<4)
	if out != want {
		t.Errorf("fixed(%#v) = %v; want %v", in, out, want)
	}
}
