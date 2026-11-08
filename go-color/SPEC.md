# Technical Specification: go-color

## 1. Domain & Purpose
`go-color` provides zero-allocation ANSI terminal styling, 24-bit TrueColor (RGB/HEX) formatting, and environment-aware terminal coloring for Go CLI tools, server logging, and background worker diagnostics.

## 2. Architecture & Design Principles
- **Standard ANSI & TrueColor Engine**: Supports 16 standard foreground/background ANSI colors along with 24-bit direct RGB (`38;2;r;g;b`) and HEX `#RRGGBB` color styling.
- **Environment Aware (NO_COLOR Compliance)**: Automatically checks `NO_COLOR` and `TERM=dumb` environment variables on startup. Disables escape codes if colors are not supported or explicitly disabled.
- **Fluent Builder API**: Chained methods (`Bold().Red().BgYellow().Sprintf(...)`) return clean formatted strings.
- **Zero Heavy Dependencies**: Pure Go standard library (`os`, `fmt`, `strconv`, `strings`, `sync/atomic`).

## 3. API Surface
- `func SetColorEnabled(enabled bool)`
- `func IsColorEnabled() bool`
- `type Style struct`
- `func New() *Style`
- `func (s *Style) Bold() *Style`
- `func (s *Style) Dim() *Style`
- `func (s *Style) Italic() *Style`
- `func (s *Style) Underline() *Style`
- `func (s *Style) Red() *Style`
- `func (s *Style) Green() *Style`
- `func (s *Style) Yellow() *Style`
- `func (s *Style) Blue() *Style`
- `func (s *Style) RGB(r, g, b uint8) *Style`
- `func (s *Style) HEX(hex string) (*Style, error)`
- `func (s *Style) Sprint(a ...any) string`
- `func (s *Style) Sprintf(format string, a ...any) string`
- `func Red(msg string) string`
- `func Green(msg string) string`
- `func Yellow(msg string) string`

## 4. Error Handling
- `ErrInvalidHEX`: Malformed hex color code.

## 5. Verification & Testing
- Unit tests verify standard ANSI escape wrapping, 24-bit RGB/HEX color generation, chained compound styles, and raw unescaped fallbacks when colors are disabled.
