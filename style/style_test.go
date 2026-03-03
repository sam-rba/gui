package style

import (
	"testing"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/math/fixed"
)

const (
	dpi     = 150.0
	epsilon = 0.5 // TODO: optimize precision
)

var (
	fontTtf = gomono.TTF
	units   = []unit.Unit{unit.Px, unit.Dp, unit.Pt, unit.Mm, unit.In, unit.Em, unit.Ex, unit.Ch}
)

func TestConvert(t *testing.T) {
	t.Parallel()

	s, err := New(Font(fontTtf), Dpi(dpi))
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Trivial
	for _, from := range units {
		for _, to := range units {
			testConvert(t, s, unit.Value{0, from}, to, 0)
		}
	}

	pxPerEm := fixed26ToFloat(s.face.Metrics().Height)
	pxPerEx := fixed26ToFloat(s.face.Metrics().XHeight)
	advance, ok := s.face.GlyphAdvance('0')
	if !ok {
		t.Fail()
	}
	pxPerCh := fixed26ToFloat(advance)

	// px → all
	px := 123.4
	in := px / dpi
	from := unit.Value{px, unit.Px}
	testConvert(t, s, from, unit.Px, px)
	testConvert(t, s, from, unit.Dp, in*160)
	testConvert(t, s, from, unit.Pt, in*72)
	testConvert(t, s, from, unit.Mm, in*25.4)
	testConvert(t, s, from, unit.In, in)
	testConvert(t, s, from, unit.Em, px/pxPerEm)
	testConvert(t, s, from, unit.Ex, px/pxPerEx)
	testConvert(t, s, from, unit.Ch, px/pxPerCh)

	// dp → all
	dp := 123.4
	in = dp / 160
	px = in * dpi
	from = unit.Value{dp, unit.Dp}
	testConvert(t, s, from, unit.Px, px)
	testConvert(t, s, from, unit.Dp, dp)
	testConvert(t, s, from, unit.Pt, in*72)
	testConvert(t, s, from, unit.Mm, in*25.4)
	testConvert(t, s, from, unit.In, in)
	testConvert(t, s, from, unit.Em, px/pxPerEm)
	testConvert(t, s, from, unit.Ex, px/pxPerEx)
	testConvert(t, s, from, unit.Ch, px/pxPerCh)

	// Pt → all
	pt := 123.4
	in = pt / 72
	px = in * dpi
	from = unit.Value{pt, unit.Pt}
	testConvert(t, s, from, unit.Px, px)
	testConvert(t, s, from, unit.Dp, in*160)
	testConvert(t, s, from, unit.Pt, pt)
	testConvert(t, s, from, unit.Mm, in*25.4)
	testConvert(t, s, from, unit.In, in)
	testConvert(t, s, from, unit.Em, px/pxPerEm)
	testConvert(t, s, from, unit.Ex, px/pxPerEx)
	testConvert(t, s, from, unit.Ch, px/pxPerCh)

	// Mm → all
	mm := 123.4
	in = mm / 25.4
	px = in * dpi
	from = unit.Value{mm, unit.Mm}
	testConvert(t, s, from, unit.Px, px)
	testConvert(t, s, from, unit.Dp, in*160)
	testConvert(t, s, from, unit.Pt, in*72)
	testConvert(t, s, from, unit.Mm, mm)
	testConvert(t, s, from, unit.In, in)
	testConvert(t, s, from, unit.Em, px/pxPerEm)
	testConvert(t, s, from, unit.Ex, px/pxPerEx)
	testConvert(t, s, from, unit.Ch, px/pxPerCh)

	// In → all
	in = 123.4
	px = in * dpi
	from = unit.Value{in, unit.In}
	testConvert(t, s, from, unit.Px, px)
	testConvert(t, s, from, unit.Dp, in*160)
	testConvert(t, s, from, unit.Pt, in*72)
	testConvert(t, s, from, unit.Mm, in*25.4)
	testConvert(t, s, from, unit.In, in)
	testConvert(t, s, from, unit.Em, px/pxPerEm)
	testConvert(t, s, from, unit.Ex, px/pxPerEx)
	testConvert(t, s, from, unit.Ch, px/pxPerCh)

	// Em → all
	em := 123.4
	px = em * pxPerEm
	in = px / dpi
	from = unit.Value{em, unit.Em}
	testConvert(t, s, from, unit.Px, px)
	testConvert(t, s, from, unit.Dp, in*160)
	testConvert(t, s, from, unit.Pt, in*72)
	testConvert(t, s, from, unit.Mm, in*25.4)
	testConvert(t, s, from, unit.In, in)
	testConvert(t, s, from, unit.Em, em)
	testConvert(t, s, from, unit.Ex, px/pxPerEx)
	testConvert(t, s, from, unit.Ch, px/pxPerCh)

	// Ex → all
	ex := 123.4
	px = ex * pxPerEx
	in = px / dpi
	from = unit.Value{ex, unit.Ex}
	testConvert(t, s, from, unit.Px, px)
	testConvert(t, s, from, unit.Dp, in*160)
	testConvert(t, s, from, unit.Pt, in*72)
	testConvert(t, s, from, unit.Mm, in*25.4)
	testConvert(t, s, from, unit.In, in)
	testConvert(t, s, from, unit.Em, px/pxPerEm)
	testConvert(t, s, from, unit.Ex, ex)
	testConvert(t, s, from, unit.Ch, px/pxPerCh)

	// Ch → all
	ch := 123.4
	px = ch * pxPerCh
	in = px / dpi
	from = unit.Value{ch, unit.Ch}
	testConvert(t, s, from, unit.Px, px)
	testConvert(t, s, from, unit.Dp, in*160)
	testConvert(t, s, from, unit.Pt, in*72)
	testConvert(t, s, from, unit.Mm, in*25.4)
	testConvert(t, s, from, unit.In, in)
	testConvert(t, s, from, unit.Em, px/pxPerEm)
	testConvert(t, s, from, unit.Ex, px/pxPerEx)
	testConvert(t, s, from, unit.Ch, ch)
}

func testConvert(t *testing.T, s *Style, from unit.Value, to unit.Unit, want float64) {
	out := s.Convert(from, to)
	if out.U != to || out.F < want-epsilon || out.F > want+epsilon {
		t.Errorf("Convert(%v, %v) = %v; want %v", from, to, out, unit.Value{want, to})
	}
}

func TestPixels(t *testing.T) {
	t.Parallel()

	s, err := New(Font(fontTtf), Dpi(dpi))
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	for _, u := range units {
		testPixels(t, s, unit.Value{0, u}, 0)
	}
	testPixels(t, s, unit.Value{123.4, unit.Px}, floatToFixed26(123.4))
	testPixels(t, s, unit.Value{123.4, unit.Dp}, floatToFixed26(123.4/160*dpi))
	testPixels(t, s, unit.Value{123.4, unit.Pt}, floatToFixed26(123.4/72*dpi))
	testPixels(t, s, unit.Value{123.4, unit.Mm}, floatToFixed26(123.4/25.4*dpi))
	testPixels(t, s, unit.Value{123.4, unit.In}, floatToFixed26(123.4*dpi))
	testPixels(t, s, unit.Value{123.4, unit.Em}, floatToFixed26(123.4).Mul(s.face.Metrics().Height))
	testPixels(t, s, unit.Value{123.4, unit.Ex}, floatToFixed26(123.4).Mul(s.face.Metrics().XHeight))
	if advance, ok := s.face.GlyphAdvance('0'); ok {
		testPixels(t, s, unit.Value{123.4, unit.Ch}, floatToFixed26(123.4).Mul(advance))
	} else {
		t.Fail()
	}
}

func testPixels(t *testing.T, s *Style, in unit.Value, want fixed.Int26_6) {
	out := s.Pixels(in)
	if out != want {
		t.Errorf("Pixels(%#v) = %v; want %v", in, out, want)
	}
}
