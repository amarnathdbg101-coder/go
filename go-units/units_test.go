package gounits

import (
	"testing"
	"time"
)

func TestParseBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1024 B", 1024},
		{"1 KiB", 1024},
		{"1 KB", 1000},
		{"10 MiB", 10 * 1024 * 1024},
		{"10 MB", 10 * 1000 * 1000},
		{"1.5 GiB", int64(1.5 * 1024 * 1024 * 1024)},
		{"64k", 64 * 1024},
	}

	for _, tt := range tests {
		actual, err := ParseBytes(tt.input)
		if err != nil {
			t.Fatalf("ParseBytes(%q) failed: %v", tt.input, err)
		}
		if actual != tt.expected {
			t.Errorf("ParseBytes(%q): expected %d, got %d", tt.input, tt.expected, actual)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	// Binary IEC
	if s := FormatBytes(1024 * 1024); s != "1.00 MiB" {
		t.Errorf("expected '1.00 MiB', got %s", s)
	}
	if s := FormatBytes(1536 * 1024 * 1024); s != "1.50 GiB" {
		t.Errorf("expected '1.50 GiB', got %s", s)
	}

	// Decimal SI
	if s := FormatBytesDecimal(1000 * 1000); s != "1.00 MB" {
		t.Errorf("expected '1.00 MB', got %s", s)
	}
}

func TestParseHumanDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"2d", 48 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"2d 4h 30m", (2*24+4)*time.Hour + 30*time.Minute},
		{"500ms", 500 * time.Millisecond},
		{"1.5h", 90 * time.Minute},
	}

	for _, tt := range tests {
		actual, err := ParseHumanDuration(tt.input)
		if err != nil {
			t.Fatalf("ParseHumanDuration(%q) failed: %v", tt.input, err)
		}
		if actual != tt.expected {
			t.Errorf("ParseHumanDuration(%q): expected %v, got %v", tt.input, tt.expected, actual)
		}
	}
}

func TestFormatHumanDuration(t *testing.T) {
	d := (2*24+4)*time.Hour + 30*time.Minute
	if s := FormatHumanDuration(d); s != "2d 4h 30m" {
		t.Errorf("expected '2d 4h 30m', got %s", s)
	}

	if s := FormatHumanDuration(250 * time.Millisecond); s != "250ms" {
		t.Errorf("expected '250ms', got %s", s)
	}
}
