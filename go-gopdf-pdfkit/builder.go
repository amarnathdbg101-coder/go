package gopdfkit

import (
	"os"

	"github.com/signintech/gopdf"
)

// PDFBuilder is a high-level fluent wrapper around *gopdf.GoPdf.
// It simplifies coordinates calculation, font switching, drawing table cells, and rendering headers.
type PDFBuilder struct {
	pdf        *gopdf.GoPdf
	format     PageFormat
	pageWidth  float64
	pageHeight float64
	curX       float64
	curY       float64
	margin     float64
}

// NewBuilder initializes a GoPdf document, sets up embedded UTF-8 fonts, and creates the first page.
func NewBuilder(format PageFormat, customHeight float64) (*PDFBuilder, error) {
	rect := GetPageRect(format, customHeight)

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *rect})
	pdf.AddPage()

	if err := SetupFonts(pdf); err != nil {
		return nil, err
	}

	margin := 24.0
	if format == FormatThermal80mm || format == FormatThermal58mm {
		margin = 8.0
	}

	return &PDFBuilder{
		pdf:        pdf,
		format:     format,
		pageWidth:  rect.W,
		pageHeight: rect.H,
		curX:       margin,
		curY:       margin,
		margin:     margin,
	}, nil
}

// Raw returns the underlying *gopdf.GoPdf instance for advanced custom operations.
func (b *PDFBuilder) Raw() *gopdf.GoPdf {
	return b.pdf
}

// PageWidth returns total width of the current document format in points.
func (b *PDFBuilder) PageWidth() float64 {
	return b.pageWidth
}

// PageHeight returns total height of current document format in points.
func (b *PDFBuilder) PageHeight() float64 {
	return b.pageHeight
}

// Margin returns standard margin.
func (b *PDFBuilder) Margin() float64 {
	return b.margin
}

// ContentWidth returns printable width excluding left and right margins.
func (b *PDFBuilder) ContentWidth() float64 {
	return b.pageWidth - (2 * b.margin)
}

// SetPosition moves the drawing cursor to (x, y).
func (b *PDFBuilder) SetPosition(x, y float64) {
	b.curX = x
	b.curY = y
	b.pdf.SetX(x)
	b.pdf.SetY(y)
}

// GetX returns current X.
func (b *PDFBuilder) GetX() float64 {
	return b.curX
}

// GetY returns current Y.
func (b *PDFBuilder) GetY() float64 {
	return b.curY
}

// SetFont sets font face and point size.
func (b *PDFBuilder) SetFont(font string, size int) error {
	return b.pdf.SetFont(font, "", size)
}

// SetTextColor sets fill color for text.
func (b *PDFBuilder) SetTextColor(r, g, bColor uint8) {
	b.pdf.SetTextColor(r, g, bColor)
}

// SetFillColor sets background fill color for shapes.
func (b *PDFBuilder) SetFillColor(r, g, bColor uint8) {
	b.pdf.SetFillColor(r, g, bColor)
}

// SetStrokeColor sets border outline color.
func (b *PDFBuilder) SetStrokeColor(r, g, bColor uint8) {
	b.pdf.SetStrokeColor(r, g, bColor)
}

// DrawText draws text at specific (x, y) coordinates.
func (b *PDFBuilder) DrawText(x, y float64, text string) error {
	b.pdf.SetX(x)
	b.pdf.SetY(y)
	return b.pdf.Text(text)
}

// DrawCell draws a bounded cell box with text alignment (gopdf.Left, gopdf.Right, gopdf.Center).
func (b *PDFBuilder) DrawCell(x, y, w, h float64, text string, align int) error {
	b.pdf.SetX(x)
	b.pdf.SetY(y)
	rect := &gopdf.Rect{W: w, H: h}
	return b.pdf.CellWithOption(rect, text, gopdf.CellOption{Align: align})
}

// DrawRect draws a rectangle with optional fill.
func (b *PDFBuilder) DrawRect(x, y, w, h float64, r, g, bColor uint8, fill bool) {
	if fill {
		b.pdf.SetFillColor(r, g, bColor)
		b.pdf.RectFromUpperLeftWithStyle(x, y, w, h, "F")
	} else {
		b.pdf.SetStrokeColor(r, g, bColor)
		b.pdf.RectFromUpperLeftWithStyle(x, y, w, h, "D")
	}
}

// DrawHorizontalLine draws a solid horizontal line across the printable width.
func (b *PDFBuilder) DrawHorizontalLine(y float64, r, g, bColor uint8) {
	b.pdf.SetStrokeColor(r, g, bColor)
	b.pdf.SetLineWidth(0.75)
	b.pdf.Line(b.margin, y, b.pageWidth-b.margin, y)
}

// DrawDashedLine draws a dashed dividing line across the printable width (ideal for thermal slips).
func (b *PDFBuilder) DrawDashedLine(y float64, r, g, bColor uint8) {
	b.pdf.SetStrokeColor(r, g, bColor)
	b.pdf.SetLineWidth(0.5)

	startX := b.margin
	endX := b.pageWidth - b.margin
	dashLen := 4.0
	gapLen := 3.0

	for x := startX; x < endX; x += dashLen + gapLen {
		x2 := x + dashLen
		if x2 > endX {
			x2 = endX
		}
		b.pdf.Line(x, y, x2, y)
	}
}

// AddPage appends a new page to the document.
func (b *PDFBuilder) AddPage() {
	b.pdf.AddPage()
	b.curX = b.margin
	b.curY = b.margin
}

// GetBytes exports the generated document as raw PDF byte buffer.
func (b *PDFBuilder) GetBytes() []byte {
	return b.pdf.GetBytesPdf()
}

// Save writes generated PDF buffer directly to a file on disk.
func (b *PDFBuilder) Save(outputPath string) error {
	data := b.GetBytes()
	return os.WriteFile(outputPath, data, 0644)
}
