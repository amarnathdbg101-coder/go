package gopdfkit

import (
	"fmt"

	"github.com/signintech/gopdf"
)

// GenerateInvoicePDF renders a modern, print-ready corporate or retail tax invoice in sub-25ms.
func GenerateInvoicePDF(inv Invoice, format PageFormat) ([]byte, error) {
	if format == "" {
		format = FormatA4
	}

	b, err := NewBuilder(format, 0)
	if err != nil {
		return nil, err
	}

	contentW := b.ContentWidth()
	margin := b.Margin()
	y := margin

	// 1. Top Header Banner
	_ = b.SetFont(FontBold, 18)
	b.SetTextColor(30, 58, 138) // Deep Blue
	_ = b.DrawText(margin, y+16, inv.CompanyName)

	_ = b.SetFont(FontBold, 18)
	b.SetTextColor(15, 23, 42) // Dark Slate
	_ = b.DrawCell(margin, y+16, contentW, 20, "TAX INVOICE", gopdf.Right)

	y += 38

	// Company Meta & Invoice Meta (Two Column Layout)
	_ = b.SetFont(FontRegular, 9)
	b.SetTextColor(100, 116, 139) // Slate Gray

	if inv.CompanyAddress != "" {
		_ = b.DrawText(margin, y+10, inv.CompanyAddress)
	}
	if inv.CompanyPhone != "" {
		_ = b.DrawText(margin, y+22, "Phone: "+inv.CompanyPhone)
	}
	if inv.CompanyGST != "" {
		_ = b.DrawText(margin, y+34, "GSTIN: "+inv.CompanyGST)
	}

	// Right Column: Invoice Info
	_ = b.SetFont(FontBold, 9)
	b.SetTextColor(30, 41, 59)
	_ = b.DrawCell(margin, y+10, contentW, 14, fmt.Sprintf("Invoice #: %s", inv.InvoiceNumber), gopdf.Right)
	_ = b.SetFont(FontRegular, 9)
	_ = b.DrawCell(margin, y+22, contentW, 14, fmt.Sprintf("Date: %s", inv.Date), gopdf.Right)
	if inv.DueDate != "" {
		_ = b.DrawCell(margin, y+34, contentW, 14, fmt.Sprintf("Due Date: %s", inv.DueDate), gopdf.Right)
	}

	y += 50
	b.DrawHorizontalLine(y, 226, 232, 240)
	y += 12

	// 2. Bill To Card
	b.DrawRect(margin, y, contentW, 44, 248, 250, 252, true) // Soft gray background
	b.DrawRect(margin, y, contentW, 44, 226, 232, 240, false) // Border

	_ = b.SetFont(FontBold, 8)
	b.SetTextColor(100, 116, 139)
	_ = b.DrawText(margin+10, y+12, "BILLED TO:")

	_ = b.SetFont(FontBold, 10)
	b.SetTextColor(15, 23, 42)
	_ = b.DrawText(margin+10, y+26, inv.CustomerName)

	_ = b.SetFont(FontRegular, 9)
	b.SetTextColor(71, 85, 105)
	custDetails := inv.CustomerPhone
	if inv.CustomerAddress != "" {
		if custDetails != "" {
			custDetails += " | " + inv.CustomerAddress
		} else {
			custDetails = inv.CustomerAddress
		}
	}
	if custDetails != "" {
		_ = b.DrawText(margin+10, y+38, custDetails)
	}

	y += 56

	// 3. Table Header
	headerH := 20.0
	b.DrawRect(margin, y, contentW, headerH, 241, 245, 249, true)

	colDescW := contentW * 0.48
	colQtyW := contentW * 0.14
	colPriceW := contentW * 0.18
	colTotalW := contentW * 0.20

	_ = b.SetFont(FontBold, 9)
	b.SetTextColor(71, 85, 105)

	curColX := margin + 8
	_ = b.DrawCell(curColX, y+4, colDescW, headerH, "ITEM DESCRIPTION", gopdf.Left)
	curColX += colDescW
	_ = b.DrawCell(curColX, y+4, colQtyW, headerH, "QTY", gopdf.Center)
	curColX += colQtyW
	_ = b.DrawCell(curColX, y+4, colPriceW, headerH, "PRICE", gopdf.Right)
	curColX += colPriceW
	_ = b.DrawCell(curColX, y+4, colTotalW-16, headerH, "TOTAL", gopdf.Right)

	y += headerH

	// 4. Table Rows
	currency := inv.Currency
	if currency == "" {
		currency = "₹"
	}

	rowH := 20.0
	_ = b.SetFont(FontRegular, 9)
	b.SetTextColor(30, 41, 59)

	for _, item := range inv.Items {
		curColX = margin + 8
		_ = b.DrawCell(curColX, y+4, colDescW, rowH, TruncateRunes(item.Description, 42), gopdf.Left)
		curColX += colDescW
		_ = b.DrawCell(curColX, y+4, colQtyW, rowH, fmt.Sprintf("%d", item.Quantity), gopdf.Center)
		curColX += colQtyW
		_ = b.DrawCell(curColX, y+4, colPriceW, rowH, fmt.Sprintf("%s %.2f", currency, item.UnitPrice), gopdf.Right)
		curColX += colPriceW
		_ = b.DrawCell(curColX, y+4, colTotalW-16, rowH, fmt.Sprintf("%s %.2f", currency, item.Total), gopdf.Right)

		y += rowH
		b.DrawHorizontalLine(y, 241, 245, 249) // Row separator
	}

	y += 16

	// 5. Totals Box (Right Aligned)
	totalsBoxW := 200.0
	totalsX := margin + contentW - totalsBoxW

	_ = b.SetFont(FontRegular, 9)
	b.SetTextColor(100, 116, 139)
	_ = b.DrawCell(totalsX, y, 100, 14, "Subtotal:", gopdf.Right)
	b.SetTextColor(30, 41, 59)
	_ = b.DrawCell(totalsX+100, y, 90, 14, fmt.Sprintf("%s %.2f", currency, inv.Subtotal), gopdf.Right)
	y += 16

	if inv.Tax > 0 {
		b.SetTextColor(100, 116, 139)
		_ = b.DrawCell(totalsX, y, 100, 14, "Tax:", gopdf.Right)
		b.SetTextColor(30, 41, 59)
		_ = b.DrawCell(totalsX+100, y, 90, 14, fmt.Sprintf("%s %.2f", currency, inv.Tax), gopdf.Right)
		y += 16
	}

	if inv.Discount > 0 {
		b.SetTextColor(100, 116, 139)
		_ = b.DrawCell(totalsX, y, 100, 14, "Discount:", gopdf.Right)
		b.SetTextColor(220, 38, 38) // Red for discount
		_ = b.DrawCell(totalsX+100, y, 90, 14, fmt.Sprintf("-%s %.2f", currency, inv.Discount), gopdf.Right)
		y += 16
	}

	// Grand Total Highlight
	b.DrawRect(totalsX, y, totalsBoxW, 26, 239, 246, 255, true) // Soft blue fill
	b.DrawRect(totalsX, y, totalsBoxW, 26, 37, 99, 235, false)  // Blue border

	_ = b.SetFont(FontBold, 11)
	b.SetTextColor(30, 58, 138)
	_ = b.DrawCell(totalsX+8, y+6, 80, 16, "Grand Total:", gopdf.Left)
	_ = b.DrawCell(totalsX+88, y+6, 104, 16, fmt.Sprintf("%s %.2f", currency, inv.GrandTotal), gopdf.Right)

	y += 40

	// 6. Notes
	if inv.Notes != "" {
		_ = b.SetFont(FontBold, 8)
		b.SetTextColor(100, 116, 139)
		_ = b.DrawText(margin, y, "NOTES & TERMS:")
		y += 12
		_ = b.SetFont(FontRegular, 8)
		b.SetTextColor(71, 85, 105)
		_ = b.DrawText(margin, y, inv.Notes)
		y += 20
	}

	// 7. Footer
	b.DrawHorizontalLine(b.PageHeight()-margin-16, 226, 232, 240)
	_ = b.SetFont(FontRegular, 8)
	b.SetTextColor(148, 163, 184)
	_ = b.DrawCell(margin, b.PageHeight()-margin-10, contentW, 12, "Thank you for your business! This is a computer-generated invoice.", gopdf.Center)

	return b.GetBytes(), nil
}
