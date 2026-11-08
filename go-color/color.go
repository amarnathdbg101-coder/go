package gocolor

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

var (
	colorEnabled atomic.Bool
)

func init() {
	// Respect https://no-color.org standard
	noColor := os.Getenv("NO_COLOR") != ""
	termDumb := os.Getenv("TERM") == "dumb"

	if noColor || termDumb {
		colorEnabled.Store(false)
	} else {
		colorEnabled.Store(true)
	}
}

// SetColorEnabled explicitly enables or disables color formatting globally.
func SetColorEnabled(enabled bool) {
	colorEnabled.Store(enabled)
}

// IsColorEnabled returns whether ANSI color formatting is active.
func IsColorEnabled() bool {
	return colorEnabled.Load()
}

// Style encapsulates formatting and color codes.
type Style struct {
	codes []string
}

// New creates an empty Style.
func New() *Style {
	return &Style{
		codes: make([]string, 0, 4),
	}
}

func (s *Style) add(code string) *Style {
	clone := &Style{
		codes: make([]string, len(s.codes), len(s.codes)+1),
	}
	copy(clone.codes, s.codes)
	clone.codes = append(clone.codes, code)
	return clone
}

// Styles
func (s *Style) Bold() *Style      { return s.add("1") }
func (s *Style) Dim() *Style       { return s.add("2") }
func (s *Style) Italic() *Style    { return s.add("3") }
func (s *Style) Underline() *Style { return s.add("4") }
func (s *Style) Reverse() *Style   { return s.add("7") }

// Foreground Standard Colors
func (s *Style) Black() *Style   { return s.add("30") }
func (s *Style) Red() *Style     { return s.add("31") }
func (s *Style) Green() *Style   { return s.add("32") }
func (s *Style) Yellow() *Style  { return s.add("33") }
func (s *Style) Blue() *Style    { return s.add("34") }
func (s *Style) Magenta() *Style { return s.add("35") }
func (s *Style) Cyan() *Style    { return s.add("36") }
func (s *Style) White() *Style   { return s.add("37") }
func (s *Style) Gray() *Style    { return s.add("90") }

// Background Standard Colors
func (s *Style) BgBlack() *Style   { return s.add("40") }
func (s *Style) BgRed() *Style     { return s.add("41") }
func (s *Style) BgGreen() *Style   { return s.add("42") }
func (s *Style) BgYellow() *Style  { return s.add("43") }
func (s *Style) BgBlue() *Style    { return s.add("44") }
func (s *Style) BgMagenta() *Style { return s.add("45") }
func (s *Style) BgCyan() *Style    { return s.add("46") }
func (s *Style) BgWhite() *Style   { return s.add("47") }

// RGB adds 24-bit TrueColor foreground color.
func (s *Style) RGB(r, g, b uint8) *Style {
	return s.add(fmt.Sprintf("38;2;%d;%d;%d", r, g, b))
}

// BgRGB adds 24-bit TrueColor background color.
func (s *Style) BgRGB(r, g, b uint8) *Style {
	return s.add(fmt.Sprintf("48;2;%d;%d;%d", r, g, b))
}

// HEX parses hex color (e.g. "#FF5733" or "00AAFF") and applies it to foreground.
func (s *Style) HEX(hex string) (*Style, error) {
	r, g, b, err := parseHex(hex)
	if err != nil {
		return nil, err
	}
	return s.RGB(r, g, b), nil
}

// Sprint formats the text with style if coloring is enabled.
func (s *Style) Sprint(a ...any) string {
	msg := fmt.Sprint(a...)
	if !IsColorEnabled() || len(s.codes) == 0 {
		return msg
	}
	prefix := fmt.Sprintf("\x1b[%sm", strings.Join(s.codes, ";"))
	return fmt.Sprintf("%s%s\x1b[0m", prefix, msg)
}

// Sprintf formats according to specifier with style if coloring is enabled.
func (s *Style) Sprintf(format string, a ...any) string {
	msg := fmt.Sprintf(format, a...)
	if !IsColorEnabled() || len(s.codes) == 0 {
		return msg
	}
	prefix := fmt.Sprintf("\x1b[%sm", strings.Join(s.codes, ";"))
	return fmt.Sprintf("%s%s\x1b[0m", prefix, msg)
}

// Convenient Package-Level Quick Helpers
func Red(msg string) string    { return New().Red().Sprint(msg) }
func Green(msg string) string  { return New().Green().Sprint(msg) }
func Yellow(msg string) string { return New().Yellow().Sprint(msg) }
func Blue(msg string) string   { return New().Blue().Sprint(msg) }
func Cyan(msg string) string   { return New().Cyan().Sprint(msg) }
func Gray(msg string) string   { return New().Gray().Sprint(msg) }
func Bold(msg string) string   { return New().Bold().Sprint(msg) }

func parseHex(hex string) (uint8, uint8, uint8, error) {
	s := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(s) != 6 {
		return 0, 0, 0, ErrInvalidHEX
	}

	val, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, ErrInvalidHEX
	}

	r := uint8((val >> 16) & 0xFF)
	g := uint8((val >> 8) & 0xFF)
	b := uint8(val & 0xFF)
	return r, g, b, nil
}
