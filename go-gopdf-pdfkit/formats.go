package gopdfkit

import "github.com/signintech/gopdf"

// PageFormat represents supported physical paper and receipt dimensions.
type PageFormat string

const (
	// FormatA4 represents standard international document paper (210mm x 297mm).
	FormatA4 PageFormat = "A4"

	// FormatA5 is the executive standard for retail counter tax invoices (148mm x 210mm).
	FormatA5 PageFormat = "A5"

	// FormatThermal80mm is for standard 80mm POS thermal roll receipt printers (width ~226pt).
	FormatThermal80mm PageFormat = "Thermal80mm"

	// FormatThermal58mm is for compact 58mm mini thermal printers (width ~164pt).
	FormatThermal58mm PageFormat = "Thermal58mm"
)

// Standard paper dimensions in points (72 points = 1 inch, 1 mm ≈ 2.83465 points).
const (
	A4WidthPoints       = 595.28
	A4HeightPoints      = 841.89
	A5WidthPoints       = 419.53
	A5HeightPoints      = 595.28
	Thermal80WidthPoints = 226.77
	Thermal58WidthPoints = 164.41
)

// GetPageRect returns a configured gopdf.Rect for the requested format and optional custom height.
func GetPageRect(format PageFormat, customHeight float64) *gopdf.Rect {
	switch format {
	case FormatA4:
		return &gopdf.Rect{W: A4WidthPoints, H: A4HeightPoints}
	case FormatA5:
		return &gopdf.Rect{W: A5WidthPoints, H: A5HeightPoints}
	case FormatThermal80mm:
		h := customHeight
		if h <= 0 {
			h = 600.0 // Default roll view
		}
		return &gopdf.Rect{W: Thermal80WidthPoints, H: h}
	case FormatThermal58mm:
		h := customHeight
		if h <= 0 {
			h = 500.0
		}
		return &gopdf.Rect{W: Thermal58WidthPoints, H: h}
	default:
		return &gopdf.Rect{W: A4WidthPoints, H: A4HeightPoints}
	}
}
