package style

import (
	"fmt"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	DefaultFontSize = 12.0 // points (1/72")
	DefaultDpi      = 72.0 // dots per inch
	DefaultHinting  = font.HintingFull
)

// Style defines the colors and font faces that widgets are drawn with.
type Style struct {
	font     *opentype.Font
	faceOpts opentype.FaceOptions
	face     font.Face // for measurements

	// TODO
}

// New returns a new Style with the given options.
// The Style should be closed after use.
func New(opts ...Option) (*Style, error) {
	o := parseOpts(opts...)

	fnt, err := opentype.Parse(o.font)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(fnt, &o.FaceOptions)
	if err != nil {
		return nil, err
	}

	return &Style{
		fnt,
		o.FaceOptions,
		face,
	}, nil
}

func (s *Style) Close() error {
	return s.face.Close()
}

// NewFace returns a new font.Face using the Style's Font and FaceOptions.
func (s *Style) NewFace() (font.Face, error) {
	return opentype.NewFace(s.font, &s.faceOpts)
}

// Convert implements the golang.org/x/exp/shiny/unit.Converter
// interface.
func (s *Style) Convert(v unit.Value, to unit.Unit) unit.Value {
	px := fixed26ToFloat(s.Pixels(v))
	var f float64
	switch to {
	case unit.Px:
		f = px
	case unit.Dp:
		f = s.pixelsToInches(px) * unit.DensityIndependentPixelsPerInch
	case unit.Pt:
		f = s.pixelsToInches(px) * unit.PointsPerInch
	case unit.Mm:
		f = s.pixelsToInches(px) * unit.MillimetresPerInch
	case unit.In:
		f = s.pixelsToInches(px)
	case unit.Em:
		f = px / fixed26ToFloat(s.face.Metrics().Height)
	case unit.Ex:
		f = px / fixed26ToFloat(s.face.Metrics().XHeight)
	case unit.Ch:
		f = px / fixed26ToFloat(s.zeroWidth())
	default:
		panic(fmt.Sprintf("unreachable: impossible %T: %v", to, to))
	}
	return unit.Value{f, to}
}

// Pixels implements the golang.org/x/exp/shiny/unit.Converter
// interface.
func (s *Style) Pixels(v unit.Value) fixed.Int26_6 {
	f := floatToFixed26(v.F)
	switch v.U {
	case unit.Px:
		return f
	case unit.Dp:
		return s.inchesToPixels(v.F / unit.DensityIndependentPixelsPerInch)
	case unit.Pt:
		return s.inchesToPixels(v.F / unit.PointsPerInch)
	case unit.Mm:
		return s.inchesToPixels(v.F / unit.MillimetresPerInch)
	case unit.In:
		return s.inchesToPixels(v.F)
	case unit.Em:
		return f.Mul(s.face.Metrics().Height)
	case unit.Ex:
		return f.Mul(s.face.Metrics().XHeight)
	case unit.Ch:
		return f.Mul(s.zeroWidth())
	default:
		panic(fmt.Sprintf("unreachable: impossible %T: %v", v.U, v.U))
	}
}

func (s *Style) pixelsToInches(px float64) (in float64) {
	return px / s.faceOpts.DPI
}

func (s *Style) inchesToPixels(in float64) (px fixed.Int26_6) {
	return floatToFixed26(in * s.faceOpts.DPI)
}

func (s *Style) zeroWidth() (px fixed.Int26_6) {
	if advance, ok := s.face.GlyphAdvance('0'); ok {
		return advance
	} else {
		return floatToFixed26(0.5).Mul(s.face.Metrics().Height) // 0.5em
	}
}
