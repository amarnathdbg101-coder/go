package gopdfkit

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/signintech/gopdf"
)

// Embed Google Noto Sans Devanagari TrueType fonts directly into the Go binary.
// This guarantees zero external font/OS dependencies and full Hindi + English + ₹ (Rupee) symbol support.
//
//go:embed fonts/NotoSansDevanagari-Regular.ttf
var fontNotoRegular []byte

//go:embed fonts/NotoSansDevanagari-Bold.ttf
var fontNotoBold []byte

const (
	FontRegular = "noto-reg"
	FontBold    = "noto-bold"
)

// SetupFonts registers the embedded TTF font family in the provided GoPdf instance.
func SetupFonts(pdf *gopdf.GoPdf) error {
	if err := pdf.AddTTFFontData(FontRegular, fontNotoRegular); err != nil {
		return fmt.Errorf("gopdfkit: failed to register regular font: %w", err)
	}
	if err := pdf.AddTTFFontData(FontBold, fontNotoBold); err != nil {
		return fmt.Errorf("gopdfkit: failed to register bold font: %w", err)
	}
	return nil
}

// TruncateRunes cleanly truncates UTF-8 strings by rune count rather than byte count,
// preventing broken multi-byte sequences for Hindi (Devanagari) and other non-ASCII scripts.
func TruncateRunes(s string, maxRunes int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}
