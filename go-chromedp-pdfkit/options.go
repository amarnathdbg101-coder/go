package pdfkit

import "time"

// Standard paper dimensions in inches.
const (
	A4WidthInches    = 8.27
	A4HeightInches   = 11.69
	LetterWidthInches = 8.50
	LetterHeightInches = 11.00
	LegalWidthInches  = 8.50
	LegalHeightInches = 14.00
)

// PDFOptions defines layout, paper sizing, margins, and rendering triggers for PDF generation.
type PDFOptions struct {
	// Paper dimensions in inches (default: A4 -> 8.27 x 11.69)
	PaperWidth  float64
	PaperHeight float64

	// Margins in inches (default: 0.4 inches)
	MarginTop    float64
	MarginBottom float64
	MarginLeft   float64
	MarginRight  float64

	// Landscape orientation (default: false -> Portrait)
	Landscape bool

	// PrintBackground prints CSS background colors, gradients, and images (Essential for Invoices/Badges)
	PrintBackground bool

	// Scale factor of the webpage rendering (default: 1.0, range: 0.1 to 2.0)
	Scale float64

	// WaitSelector optionally halts PDF print until a specific CSS DOM element is rendered (e.g. "#invoice-ready")
	WaitSelector string

	// WaitDuration optionally waits for fonts, charts, or images to stabilize before printing
	WaitDuration time.Duration

	// Header and Footer HTML templates
	HeaderTemplate      string
	FooterTemplate      string
	DisplayHeaderFooter bool
}

// DefaultOptions returns standard A4 portrait settings with background printing enabled.
func DefaultOptions() PDFOptions {
	return PDFOptions{
		PaperWidth:      A4WidthInches,
		PaperHeight:     A4HeightInches,
		MarginTop:       0.4,
		MarginBottom:    0.4,
		MarginLeft:      0.4,
		MarginRight:     0.4,
		Landscape:       false,
		PrintBackground: true,
		Scale:           1.0,
	}
}

// InvoiceOptions returns optimized layout options for A4 retail invoices and bills.
func InvoiceOptions() PDFOptions {
	return PDFOptions{
		PaperWidth:      A4WidthInches,
		PaperHeight:     A4HeightInches,
		MarginTop:       0.3,
		MarginBottom:    0.3,
		MarginLeft:      0.3,
		MarginRight:     0.3,
		Landscape:       false,
		PrintBackground: true,
		Scale:           1.0,
	}
}

// BarcodeLabelOptions returns layout options for thermal barcode/shipping label printers.
// Standard thermal label sizes: 4.0 x 6.0 inches or 2.0 x 1.0 inches.
func BarcodeLabelOptions(widthInches, heightInches float64) PDFOptions {
	if widthInches <= 0 {
		widthInches = 4.0
	}
	if heightInches <= 0 {
		heightInches = 6.0
	}
	return PDFOptions{
		PaperWidth:      widthInches,
		PaperHeight:     heightInches,
		MarginTop:       0.0,
		MarginBottom:    0.0,
		MarginLeft:      0.0,
		MarginRight:     0.0,
		Landscape:       false,
		PrintBackground: true,
		Scale:           1.0,
	}
}

// Sanitize applies sane bounds to options.
func (o *PDFOptions) Sanitize() {
	if o.PaperWidth <= 0 {
		o.PaperWidth = A4WidthInches
	}
	if o.PaperHeight <= 0 {
		o.PaperHeight = A4HeightInches
	}
	if o.Scale <= 0 || o.Scale > 2.0 {
		o.Scale = 1.0
	}
}
