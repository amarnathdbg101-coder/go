package gopdfkit

import (
	"fmt"

	"github.com/signintech/gopdf"
)

// GenerateReceiptPDF renders a compact, high-speed thermal roll receipt for 80mm or 58mm POS printers.
func GenerateReceiptPDF(rec Receipt, format PageFormat) ([]byte, error) {
	if format != FormatThermal58mm {
		format = FormatThermal80mm
	}

	// Calculate dynamic height based on item count so paper roll doesn't waste space
	itemCount := len(rec.Items)
	baseHeight := 180.0
	calculatedHeight := baseHeight + float64(itemCount*18)
	if calculatedHeight < 350.0 {
		calculatedHeight = 350.0
	}

	b, err := NewBuilder(format, calculatedHeight)
	if err != nil {
		return nil, err
	}

	contentW := b.ContentWidth()
	margin := b.Margin()
	y := margin

	// 1. Store Header (Centered)
	_ = b.SetFont(FontBold, 12)
	b.SetTextColor(0, 0, 0)
	_ = b.DrawCell(margin, y, contentW, 16, rec.StoreName, gopdf.Center)
	y += 18

	_ = b.SetFont(FontRegular, 8)
	b.SetTextColor(60, 60, 60)
	if rec.StoreAddress != "" {
		_ = b.DrawCell(margin, y, contentW, 12, rec.StoreAddress, gopdf.Center)
		y += 13
	}
	if rec.StorePhone != "" {
		_ = b.DrawCell(margin, y, contentW, 12, "Tel: "+rec.StorePhone, gopdf.Center)
		y += 13
	}

	// 2. Receipt Meta
	b.DrawDashedLine(y+4, 150, 150, 150)
	y += 10

	_ = b.SetFont(FontRegular, 8)
	_ = b.DrawCell(margin, y, contentW*0.5, 12, fmt.Sprintf("Receipt: %s", rec.ReceiptNo), gopdf.Left)
	_ = b.DrawCell(margin+contentW*0.5, y, contentW*0.5, 12, rec.Date, gopdf.Right)
	y += 14

	if rec.Cashier != "" {
		_ = b.DrawCell(margin, y, contentW, 12, "Cashier: "+rec.Cashier, gopdf.Left)
		y += 14
	}

	b.DrawDashedLine(y+2, 150, 150, 150)
	y += 8

	// 3. Items Header
	colItemW := contentW * 0.52
	colQtyW := contentW * 0.18
	colTotalW := contentW * 0.30

	_ = b.SetFont(FontBold, 8)
	_ = b.DrawCell(margin, y, colItemW, 12, "ITEM", gopdf.Left)
	_ = b.DrawCell(margin+colItemW, y, colQtyW, 12, "QTY", gopdf.Center)
	_ = b.DrawCell(margin+colItemW+colQtyW, y, colTotalW, 12, "TOTAL", gopdf.Right)
	y += 14

	b.DrawDashedLine(y, 180, 180, 180)
	y += 6

	// 4. Line Items
	currency := rec.Currency
	if currency == "" {
		currency = "₹"
	}

	_ = b.SetFont(FontRegular, 8)
	for _, item := range rec.Items {
		_ = b.DrawCell(margin, y, colItemW, 12, TruncateRunes(item.Name, 18), gopdf.Left)
		_ = b.DrawCell(margin+colItemW, y, colQtyW, 12, fmt.Sprintf("%d", item.Qty), gopdf.Center)
		_ = b.DrawCell(margin+colItemW+colQtyW, y, colTotalW, 12, fmt.Sprintf("%s%.2f", currency, item.Amount), gopdf.Right)
		y += 14
	}

	b.DrawDashedLine(y+2, 150, 150, 150)
	y += 10

	// 5. Totals
	_ = b.SetFont(FontRegular, 8)
	_ = b.DrawCell(margin, y, contentW*0.5, 12, "Subtotal:", gopdf.Left)
	_ = b.DrawCell(margin+contentW*0.5, y, contentW*0.5, 12, fmt.Sprintf("%s%.2f", currency, rec.Subtotal), gopdf.Right)
	y += 13

	if rec.Tax > 0 {
		_ = b.DrawCell(margin, y, contentW*0.5, 12, "Tax:", gopdf.Left)
		_ = b.DrawCell(margin+contentW*0.5, y, contentW*0.5, 12, fmt.Sprintf("%s%.2f", currency, rec.Tax), gopdf.Right)
		y += 13
	}

	// Grand Total (Bold)
	_ = b.SetFont(FontBold, 10)
	_ = b.DrawCell(margin, y+2, contentW*0.5, 14, "TOTAL:", gopdf.Left)
	_ = b.DrawCell(margin+contentW*0.5, y+2, contentW*0.5, 14, fmt.Sprintf("%s%.2f", currency, rec.GrandTotal), gopdf.Right)
	y += 20

	if rec.PaymentMode != "" {
		_ = b.SetFont(FontRegular, 8)
		_ = b.DrawCell(margin, y, contentW, 12, "Payment Mode: "+rec.PaymentMode, gopdf.Center)
		y += 14
	}

	// 6. Thermal Footer
	b.DrawDashedLine(y+2, 150, 150, 150)
	y += 10

	_ = b.SetFont(FontRegular, 7)
	b.SetTextColor(80, 80, 80)
	footerMsg := rec.FooterNote
	if footerMsg == "" {
		footerMsg = "Thank you! Please visit again."
	}
	_ = b.DrawCell(margin, y, contentW, 12, footerMsg, gopdf.Center)

	return b.GetBytes(), nil
}
