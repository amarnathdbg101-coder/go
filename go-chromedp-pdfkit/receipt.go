package pdfkit

import (
	"context"
)

// ReceiptItem represents an item in a POS thermal receipt.
type ReceiptItem struct {
	Name   string  `json:"name"`
	Qty    int     `json:"qty"`
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
}

// ReceiptData encapsulates all fields required to print a 58mm/80mm POS receipt.
type ReceiptData struct {
	ShopName       string        `json:"shop_name"`
	ShopAddress    string        `json:"shop_address"`
	ShopPhone      string        `json:"shop_phone"`
	GSTIN          string        `json:"gstin,omitempty"`
	BillNumber     string        `json:"bill_number"`
	Date           string        `json:"date"`
	CustomerName   string        `json:"customer_name,omitempty"`
	CustomerPhone  string        `json:"customer_phone,omitempty"`
	Items          []ReceiptItem `json:"items"`
	Subtotal       float64       `json:"subtotal"`
	TaxAmount      float64       `json:"tax_amount"`
	DiscountAmount float64       `json:"discount_amount,omitempty"`
	TotalAmount    float64       `json:"total_amount"`
	PaymentMethod  string        `json:"payment_method"` // e.g. "Cash", "UPI", "Card"
	FooterNote     string        `json:"footer_note,omitempty"`
}

// GenerateThermalReceipt generates a POS thermal receipt PDF (80mm width by default).
func GenerateThermalReceipt(ctx context.Context, data ReceiptData, widthMm float64) ([]byte, error) {
	if widthMm <= 0 {
		widthMm = 80 // Default 80mm POS thermal paper
	}
	widthInches := widthMm / 25.4
	heightInches := 11.0 // Continuous thermal roll length

	opts := PDFOptions{
		PaperWidth:      widthInches,
		PaperHeight:     heightInches,
		MarginTop:       0.1,
		MarginBottom:    0.1,
		MarginLeft:      0.1,
		MarginRight:     0.1,
		PrintBackground: true,
		Scale:           1.0,
	}

	return defaultEngine.GenerateFromTemplate(ctx, DefaultReceiptTemplate, data, opts)
}

// DefaultReceiptTemplate is a high-contrast, thermal-printer-optimized HTML/CSS receipt template.
const DefaultReceiptTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  @page { margin: 0; }
  body {
    font-family: "Courier New", Courier, monospace, monospace;
    font-size: 12px;
    color: #000000;
    margin: 0;
    padding: 8px;
    background: #ffffff;
    box-sizing: border-box;
  }
  .text-center { text-align: center; }
  .text-right { text-align: right; }
  .bold { font-weight: bold; }
  .divider {
    border-top: 1px dashed #000000;
    margin: 6px 0;
  }
  .double-divider {
    border-top: 2px solid #000000;
    margin: 6px 0;
  }
  .shop-title {
    font-size: 16px;
    font-weight: bold;
    text-transform: uppercase;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    margin: 4px 0;
  }
  td, th {
    padding: 2px 0;
    font-size: 11px;
  }
  th { text-align: left; }
</style>
</head>
<body>
  <div class="text-center">
    <div class="shop-title">{{.ShopName}}</div>
    <div>{{.ShopAddress}}</div>
    <div>Ph: {{.ShopPhone}}</div>
    {{if .GSTIN}}<div>GSTIN: {{.GSTIN}}</div>{{end}}
  </div>

  <div class="divider"></div>

  <div>
    <div>Bill No: <span class="bold">{{.BillNumber}}</span></div>
    <div>Date: {{.Date}}</div>
    {{if .CustomerName}}<div>Cust: {{.CustomerName}} {{if .CustomerPhone}}({{.CustomerPhone}}){{end}}</div>{{end}}
  </div>

  <div class="divider"></div>

  <table>
    <thead>
      <tr>
        <th style="width: 50%;">ITEM</th>
        <th class="text-center" style="width: 15%;">QTY</th>
        <th class="text-right" style="width: 35%;">AMT</th>
      </tr>
    </thead>
    <tbody>
      {{range .Items}}
      <tr>
        <td>{{.Name}}</td>
        <td class="text-center">{{.Qty}}</td>
        <td class="text-right">{{printf "%.2f" .Amount}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>

  <div class="divider"></div>

  <table>
    <tr>
      <td>Subtotal:</td>
      <td class="text-right">{{printf "%.2f" .Subtotal}}</td>
    </tr>
    {{if gt .TaxAmount 0.0}}
    <tr>
      <td>Tax/GST:</td>
      <td class="text-right">{{printf "%.2f" .TaxAmount}}</td>
    </tr>
    {{end}}
    {{if gt .DiscountAmount 0.0}}
    <tr>
      <td>Discount:</td>
      <td class="text-right">-{{printf "%.2f" .DiscountAmount}}</td>
    </tr>
    {{end}}
    <tr class="bold" style="font-size: 13px;">
      <td>TOTAL:</td>
      <td class="text-right">₹{{printf "%.2f" .TotalAmount}}</td>
    </tr>
  </table>

  <div class="divider"></div>

  <div>Payment Mode: <span class="bold">{{.PaymentMethod}}</span></div>

  <div class="divider"></div>

  <div class="text-center" style="font-size: 10px; margin-top: 8px;">
    {{if .FooterNote}}{{.FooterNote}}{{else}}Thank you for shopping with us! Visit again.{{end}}
  </div>
</body>
</html>`
