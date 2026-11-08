package goversion

import (
	"testing"
)

func TestVersion_ParsingAndString(t *testing.T) {
	v, err := Parse("v1.2.3-beta.1+build.42")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("version mismatch: %d.%d.%d", v.Major, v.Minor, v.Patch)
	}
	if v.Prerelease != "beta.1" {
		t.Errorf("prerelease mismatch: %s", v.Prerelease)
	}
	if v.BuildMetadata != "build.42" {
		t.Errorf("build metadata mismatch: %s", v.BuildMetadata)
	}

	if str := v.String(); str != "1.2.3-beta.1+build.42" {
		t.Errorf("formatted string mismatch: %s", str)
	}
}

func TestVersion_SemVerPrecedence(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int // -1: v1 < v2, 0: v1 == v2, 1: v1 > v2
	}{
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.9.9", 1},
		{"1.2.0", "1.2.1", -1},
		{"1.2.3", "1.2.3", 0},
		// Normal version has higher precedence than prerelease
		{"1.0.0", "1.0.0-rc.1", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.2", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.beta", -1}, // numeric < non-numeric
		// Build metadata ignored in precedence
		{"1.0.0+build.1", "1.0.0+build.2", 0},
	}

	for _, tt := range tests {
		ver1 := MustParse(tt.v1)
		ver2 := MustParse(tt.v2)

		cmp := ver1.Compare(ver2)
		if cmp != tt.expected {
			t.Errorf("Compare(%s, %s): expected %d, got %d", tt.v1, tt.v2, tt.expected, cmp)
		}
	}
}

func TestConstraint_Ranges(t *testing.T) {
	// Compound range
	c, err := ParseConstraint(">= 1.0.0, < 2.0.0")
	if err != nil {
		t.Fatalf("ParseConstraint failed: %v", err)
	}

	if !c.CheckString("1.5.0") {
		t.Errorf("expected 1.5.0 to satisfy '>= 1.0.0, < 2.0.0'")
	}
	if c.CheckString("2.0.0") {
		t.Errorf("expected 2.0.0 to fail '< 2.0.0'")
	}
	if c.CheckString("0.9.0") {
		t.Errorf("expected 0.9.0 to fail '>= 1.0.0'")
	}

	// Caret range: ^1.2.3 (matches >= 1.2.3, < 2.0.0)
	caret, err := ParseConstraint("^1.2.3")
	if err != nil {
		t.Fatalf("Parse Caret failed: %v", err)
	}
	if !caret.CheckString("1.9.9") {
		t.Errorf("expected 1.9.9 to satisfy ^1.2.3")
	}
	if caret.CheckString("1.2.2") {
		t.Errorf("expected 1.2.2 to fail ^1.2.3")
	}
	if caret.CheckString("2.0.0") {
		t.Errorf("expected 2.0.0 to fail ^1.2.3")
	}

	// Tilde range: ~1.2.3 (matches >= 1.2.3, < 1.3.0)
	tilde, err := ParseConstraint("~1.2.3")
	if err != nil {
		t.Fatalf("Parse Tilde failed: %v", err)
	}
	if !tilde.CheckString("1.2.9") {
		t.Errorf("expected 1.2.9 to satisfy ~1.2.3")
	}
	if tilde.CheckString("1.3.0") {
		t.Errorf("expected 1.3.0 to fail ~1.2.3")
	}
}
