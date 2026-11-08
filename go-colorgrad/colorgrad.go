package colorgrad

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Color represents an 8-bit RGB color.
type Color struct {
	R uint8
	G uint8
	B uint8
}

// RGB creates a new Color from red, green, and blue components.
func RGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b}
}

// Hex parses a hexadecimal color string (e.g., "#FF0000", "00FF00", "#F00").
func Hex(s string) (Color, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) == 3 {
		// Expand shorthand: "F0A" -> "FF00AA"
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return Color{}, ErrInvalidHex
	}

	val, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return Color{}, ErrInvalidHex
	}

	return Color{
		R: uint8(val >> 16),
		G: uint8((val >> 8) & 0xFF),
		B: uint8(val & 0xFF),
	}, nil
}

// Hex returns the uppercase 6-character hex representation with leading '#'.
func (c Color) Hex() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// Stop marks a color at a relative position [0.0, 1.0].
type Stop struct {
	Pos   float64
	Color Color
}

// Gradient represents an interpolated sequence of color stops.
type Gradient struct {
	stops []Stop
}

// New creates a gradient with colors evenly spaced from 0.0 to 1.0.
func New(colors ...Color) (*Gradient, error) {
	if len(colors) < 2 {
		return nil, ErrInsufficientStops
	}

	stops := make([]Stop, len(colors))
	step := 1.0 / float64(len(colors)-1)

	for i, c := range colors {
		pos := float64(i) * step
		if i == len(colors)-1 {
			pos = 1.0
		}
		stops[i] = Stop{Pos: pos, Color: c}
	}

	return &Gradient{stops: stops}, nil
}

// NewFromHex creates a gradient from hex color strings.
func NewFromHex(hexes ...string) (*Gradient, error) {
	if len(hexes) < 2 {
		return nil, ErrInsufficientStops
	}

	colors := make([]Color, len(hexes))
	for i, h := range hexes {
		c, err := Hex(h)
		if err != nil {
			return nil, err
		}
		colors[i] = c
	}

	return New(colors...)
}

// At interpolates the color at position t (clamped between 0.0 and 1.0).
func (g *Gradient) At(t float64) Color {
	if t <= 0.0 {
		return g.stops[0].Color
	}
	if t >= 1.0 {
		return g.stops[len(g.stops)-1].Color
	}

	// Find the bounding stops
	idx := sort.Search(len(g.stops), func(i int) bool {
		return g.stops[i].Pos >= t
	})

	if idx == 0 {
		return g.stops[0].Color
	}

	lower := g.stops[idx-1]
	upper := g.stops[idx]

	span := upper.Pos - lower.Pos
	if span <= 0 {
		return lower.Color
	}

	factor := (t - lower.Pos) / span

	// Linear interpolation in RGB space
	r := uint8(math.Round(float64(lower.Color.R) + factor*(float64(upper.Color.R)-float64(lower.Color.R))))
	green := uint8(math.Round(float64(lower.Color.G) + factor*(float64(upper.Color.G)-float64(lower.Color.G))))
	b := uint8(math.Round(float64(lower.Color.B) + factor*(float64(upper.Color.B)-float64(lower.Color.B))))

	return Color{R: r, G: green, B: b}
}

// Ramp returns a slice of N evenly interpolated colors.
func (g *Gradient) Ramp(steps int) []Color {
	if steps <= 0 {
		return nil
	}
	if steps == 1 {
		return []Color{g.At(0.5)}
	}

	colors := make([]Color, steps)
	stepSize := 1.0 / float64(steps-1)
	for i := 0; i < steps; i++ {
		t := float64(i) * stepSize
		if i == steps-1 {
			t = 1.0
		}
		colors[i] = g.At(t)
	}
	return colors
}

// IsNoColor checks if terminal colors should be disabled per NO_COLOR specification.
func IsNoColor() bool {
	return os.Getenv("NO_COLOR") != ""
}

// Style wraps text in 24-bit TrueColor ANSI escape codes.
func Style(text string, c Color) string {
	if IsNoColor() || text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", c.R, c.G, c.B, text)
}

// Apply applies a smooth color gradient across each rune in the text.
func (g *Gradient) Apply(text string) string {
	if IsNoColor() || text == "" {
		return text
	}

	runes := []rune(text)
	if len(runes) == 1 {
		return Style(text, g.At(0.5))
	}

	var sb strings.Builder
	step := 1.0 / float64(len(runes)-1)

	for i, r := range runes {
		t := float64(i) * step
		if i == len(runes)-1 {
			t = 1.0
		}
		c := g.At(t)
		sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c\x1b[0m", c.R, c.G, c.B, r))
	}

	return sb.String()
}

// ProgressBar renders a visual terminal progress bar with dynamic gradient coloring.
func ProgressBar(percent float64, width int, g *Gradient) string {
	if width <= 0 {
		width = 20
	}
	if percent < 0.0 {
		percent = 0.0
	}
	if percent > 1.0 {
		percent = 1.0
	}

	filledLen := int(math.Round(percent * float64(width)))
	if filledLen > width {
		filledLen = width
	}

	c := g.At(percent)
	filled := strings.Repeat("█", filledLen)
	empty := strings.Repeat("░", width-filledLen)

	if IsNoColor() {
		return fmt.Sprintf("[%s%s] %3.0f%%", filled, empty, percent*100)
	}

	coloredFilled := fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", c.R, c.G, c.B, filled)
	return fmt.Sprintf("[%s%s] %3.0f%%", coloredFilled, empty, percent*100)
}

// Preset Gradients:

// TrafficLight creates a Red -> Yellow -> Green gradient (ideal for health and progress).
func TrafficLight() *Gradient {
	g, _ := NewFromHex("#E53935", "#FDD835", "#43A047")
	return g
}

// Heat creates a Black -> Red -> Orange -> Yellow -> White heat map.
func Heat() *Gradient {
	g, _ := NewFromHex("#000000", "#D32F2F", "#FF9800", "#FFEB3B", "#FFFFFF")
	return g
}

// Sunset creates a Purple -> Pink -> Orange -> Gold sunset gradient.
func Sunset() *Gradient {
	g, _ := NewFromHex("#311B92", "#AD1457", "#E65100", "#FFD54F")
	return g
}

// Neon creates a Cyberpunk Cyan -> Magenta -> Purple neon glow.
func Neon() *Gradient {
	g, _ := NewFromHex("#00F0FF", "#FF007F", "#7928CA")
	return g
}

// Rainbow creates a full 7-color ROYGBIV spectrum.
func Rainbow() *Gradient {
	g, _ := NewFromHex("#FF0000", "#FF7F00", "#FFFF00", "#00FF00", "#0000FF", "#4B0082", "#9400D3")
	return g
}
