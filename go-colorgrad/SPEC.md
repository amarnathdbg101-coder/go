# SPEC: go-colorgrad

## 1. Metadata
- **Module**: `go-colorgrad`
- **Package**: `colorgrad`
- **Location**: `C:\Library\go-colorgrad`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `fmt`, `math`, `os`, `sort`, `strconv`, `strings`).

---

## 2. Core Capabilities
1. **Linear Color Space Interpolation**:
   - Computes linear RGB color stops across relative intervals $[0.0, 1.0]$ with continuous `At(t)` interpolation.
2. **24-bit TrueColor Terminal Rendering**:
   - Outputs standard ANSI escape sequences (`\x1b[38;2;R;G;Bm`) for TrueColor terminals.
   - Smooth multi-rune gradient text rendering via `Apply(text)`.
3. **Progress Bar Visualizer**:
   - `ProgressBar(percent, width, gradient)` renders visual terminal progress bars with dynamic status colors.
4. **NO_COLOR Standard Compliance**:
   - Automatically bypasses escape sequences when the `NO_COLOR` environment variable is present.
5. **Built-in Presets**:
   - `TrafficLight()` (Red-Yellow-Green), `Heat()`, `Sunset()`, `Neon()`, `Rainbow()`.

---

## 3. Public API
```go
type Color struct {
    R, G, B uint8
}
func RGB(r, g, b uint8) Color
func Hex(s string) (Color, error)
func (c Color) Hex() string

type Stop struct {
    Pos   float64
    Color Color
}

type Gradient struct
func New(colors ...Color) (*Gradient, error)
func NewFromHex(hexes ...string) (*Gradient, error)
func (g *Gradient) At(t float64) Color
func (g *Gradient) Ramp(steps int) []Color
func (g *Gradient) Apply(text string) string

func Style(text string, c Color) string
func ProgressBar(percent float64, width int, g *Gradient) string
func IsNoColor() bool

func TrafficLight() *Gradient
func Heat() *Gradient
func Sunset() *Gradient
func Neon() *Gradient
func Rainbow() *Gradient
```
