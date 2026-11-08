package colorgrad

import (
	"os"
	"strings"
	"testing"
)

func TestHexParsing(t *testing.T) {
	c, err := Hex("#FF00AA")
	if err != nil {
		t.Fatalf("failed to parse hex: %v", err)
	}
	if c.R != 255 || c.G != 0 || c.B != 170 {
		t.Fatalf("unexpected color: %+v", c)
	}
	if c.Hex() != "#FF00AA" {
		t.Fatalf("expected #FF00AA, got %s", c.Hex())
	}

	// Shorthand
	cShort, err := Hex("#F0A")
	if err != nil {
		t.Fatalf("failed to parse shorthand: %v", err)
	}
	if cShort != c {
		t.Fatalf("shorthand did not match expanded hex")
	}

	// Invalid
	if _, err := Hex("ZZZZZZ"); err == nil {
		t.Fatalf("expected error on invalid hex")
	}
}

func TestGradient_Interpolation(t *testing.T) {
	// Black to White
	black := RGB(0, 0, 0)
	white := RGB(255, 255, 255)
	g, err := New(black, white)
	if err != nil {
		t.Fatalf("failed to create gradient: %v", err)
	}

	// At 0.0 -> Black
	if g.At(0.0) != black {
		t.Fatalf("expected black at 0.0")
	}
	// At 1.0 -> White
	if g.At(1.0) != white {
		t.Fatalf("expected white at 1.0")
	}
	// At 0.5 -> Mid Gray (~128, 128, 128)
	mid := g.At(0.5)
	if mid.R != 128 || mid.G != 128 || mid.B != 128 {
		t.Fatalf("expected mid gray (128,128,128), got %+v", mid)
	}

	// Ramp of 5
	ramp := g.Ramp(5)
	if len(ramp) != 5 {
		t.Fatalf("expected ramp of 5, got %d", len(ramp))
	}
	if ramp[0] != black || ramp[4] != white {
		t.Fatalf("ramp endpoints mismatch")
	}
}

func TestGradient_TextAndProgress(t *testing.T) {
	_ = os.Unsetenv("NO_COLOR")

	g := TrafficLight()
	styled := g.Apply("Antigravity")
	if !strings.Contains(styled, "\x1b[38;2;") {
		t.Fatalf("expected ANSI TrueColor escape sequences")
	}

	bar := ProgressBar(0.75, 20, g)
	if !strings.Contains(bar, "75%") {
		t.Fatalf("expected 75%% in progress bar, got: %s", bar)
	}

	// Test NO_COLOR compliance
	_ = os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	noColorText := g.Apply("Hello")
	if strings.Contains(noColorText, "\x1b") {
		t.Fatalf("expected no escape codes when NO_COLOR is set")
	}

	noColorBar := ProgressBar(0.50, 20, g)
	if strings.Contains(noColorBar, "\x1b") {
		t.Fatalf("expected no escape codes in progress bar when NO_COLOR is set")
	}
}

func TestPresets(t *testing.T) {
	presets := []*Gradient{
		TrafficLight(),
		Heat(),
		Sunset(),
		Neon(),
		Rainbow(),
	}

	for i, p := range presets {
		if p == nil || len(p.stops) < 2 {
			t.Fatalf("preset %d is invalid", i)
		}
	}
}
