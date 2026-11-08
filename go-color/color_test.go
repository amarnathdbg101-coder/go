package gocolor

import (
	"strings"
	"testing"
)

func TestColor_Formatting(t *testing.T) {
	SetColorEnabled(true)

	redText := Red("error")
	if !strings.HasPrefix(redText, "\x1b[31m") || !strings.HasSuffix(redText, "\x1b[0m") {
		t.Errorf("expected red ANSI wrapping, got %q", redText)
	}

	greenText := Green("success")
	if !strings.HasPrefix(greenText, "\x1b[32m") || !strings.HasSuffix(greenText, "\x1b[0m") {
		t.Errorf("expected green ANSI wrapping, got %q", greenText)
	}

	chained := New().Bold().Red().BgYellow().Sprint("alert")
	expectedPrefix := "\x1b[1;31;43m"
	if !strings.HasPrefix(chained, expectedPrefix) {
		t.Errorf("expected prefix %q, got %q", expectedPrefix, chained)
	}
}

func TestColor_TrueColorRGBAndHEX(t *testing.T) {
	SetColorEnabled(true)

	// RGB (255, 100, 50)
	rgbText := New().RGB(255, 100, 50).Sprint("orange")
	expectedPrefix := "\x1b[38;2;255;100;50m"
	if !strings.HasPrefix(rgbText, expectedPrefix) {
		t.Errorf("expected RGB prefix %q, got %q", expectedPrefix, rgbText)
	}

	// HEX #00FF00
	hexStyle, err := New().HEX("#00FF00")
	if err != nil {
		t.Fatalf("HEX parse failed: %v", err)
	}
	hexText := hexStyle.Sprint("lime")
	expectedHexPrefix := "\x1b[38;2;0;255;0m"
	if !strings.HasPrefix(hexText, expectedHexPrefix) {
		t.Errorf("expected HEX prefix %q, got %q", expectedHexPrefix, hexText)
	}

	// Invalid HEX
	_, err = New().HEX("invalid")
	if err != ErrInvalidHEX {
		t.Errorf("expected ErrInvalidHEX, got %v", err)
	}
}

func TestColor_DisabledMode(t *testing.T) {
	SetColorEnabled(false)
	defer SetColorEnabled(true)

	text := Red("plain text")
	if strings.Contains(text, "\x1b[") {
		t.Errorf("expected raw unescaped text when disabled, got %q", text)
	}
	if text != "plain text" {
		t.Errorf("expected 'plain text', got %q", text)
	}
}
