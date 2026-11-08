package pdfkit

import (
	"bytes"
	"fmt"
	"html/template"
)

// RenderHTML evaluates a Go html/template string with provided data context.
func RenderHTML(tmplStr string, data any) (string, error) {
	tmpl, err := template.New("pdf_template").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateRender, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateRender, err)
	}

	return buf.String(), nil
}

// InvoiceItem represents an individual line item in an invoice.
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

// InvoiceData encapsulates common e-commerce or billing document fields.
type InvoiceData struct {
	InvoiceNumber string        `json:"invoice_number"`
	Date          string        `json:"date"`
	DueDate       string        `json:"due_date,omitempty"`
	CompanyName   string        `json:"company_name"`
	CompanyEmail  string        `json:"company_email"`
	CompanyPhone  string        `json:"company_phone"`
	CustomerName  string        `json:"customer_name"`
	CustomerEmail string        `json:"customer_email"`
	Items         []InvoiceItem `json:"items"`
	Subtotal      float64       `json:"subtotal"`
	Tax           float64       `json:"tax"`
	TotalAmount   float64       `json:"total_amount"`
	Currency      string        `json:"currency"`
	Notes         string        `json:"notes,omitempty"`
}

// DefaultInvoiceTemplate returns a responsive, print-optimized HTML/CSS template for retail and B2B invoices.
const DefaultInvoiceTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    color: #333333;
    margin: 0;
    padding: 24px;
    background-color: #ffffff;
    box-sizing: border-box;
  }
  .header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    border-bottom: 2px solid #2563eb;
    padding-bottom: 16px;
    margin-bottom: 24px;
  }
  .company-title {
    font-size: 24px;
    font-weight: 700;
    color: #1e3a8a;
    margin: 0;
  }
  .company-meta {
    font-size: 13px;
    color: #64748b;
    margin-top: 4px;
  }
  .invoice-badge {
    text-align: right;
  }
  .invoice-badge h1 {
    font-size: 26px;
    font-weight: 800;
    color: #0f172a;
    margin: 0;
    letter-spacing: -0.5px;
  }
  .invoice-meta {
    font-size: 13px;
    color: #475569;
    margin-top: 4px;
  }
  .bill-to {
    background-color: #f8fafc;
    border-radius: 8px;
    padding: 12px 16px;
    margin-bottom: 24px;
  }
  .bill-to h4 {
    margin: 0 0 6px 0;
    font-size: 12px;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .bill-to p {
    margin: 2px 0;
    font-size: 14px;
    font-weight: 600;
    color: #1e293b;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 24px;
  }
  th {
    background-color: #f1f5f9;
    color: #475569;
    font-size: 12px;
    font-weight: 700;
    text-align: left;
    padding: 10px 12px;
    border-bottom: 1px solid #cbd5e1;
    text-transform: uppercase;
  }
  td {
    padding: 10px 12px;
    font-size: 13px;
    border-bottom: 1px solid #e2e8f0;
    color: #1e293b;
  }
  .text-right { text-align: right; }
  .text-center { text-align: center; }
  .totals-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
  }
  .totals-table {
    width: 280px;
    margin-bottom: 0;
  }
  .totals-table td {
    padding: 6px 12px;
    border: none;
  }
  .grand-total {
    font-size: 16px;
    font-weight: 800;
    color: #1e3a8a;
    border-top: 2px solid #2563eb !important;
  }
  .footer {
    margin-top: 40px;
    padding-top: 16px;
    border-top: 1px solid #e2e8f0;
    font-size: 11px;
    color: #94a3b8;
    text-align: center;
  }
</style>
</head>
<body>
  <div class="header">
    <div>
      <h2 class="company-title">{{.CompanyName}}</h2>
      <div class="company-meta">{{.CompanyEmail}} | {{.CompanyPhone}}</div>
    </div>
    <div class="invoice-badge">
      <h1>INVOICE</h1>
      <div class="invoice-meta">#{{.InvoiceNumber}} &bull; Date: {{.Date}}</div>
    </div>
  </div>

  <div class="bill-to">
    <h4>Billed To:</h4>
    <p>{{.CustomerName}}</p>
    <div style="font-size: 13px; color: #64748b;">{{.CustomerEmail}}</div>
  </div>

  <table>
    <thead>
      <tr>
        <th style="width: 50%;">Description</th>
        <th class="text-center" style="width: 15%;">Qty</th>
        <th class="text-right" style="width: 15%;">Price</th>
        <th class="text-right" style="width: 20%;">Total</th>
      </tr>
    </thead>
    <tbody>
      {{range .Items}}
      <tr>
        <td>{{.Description}}</td>
        <td class="text-center">{{.Quantity}}</td>
        <td class="text-right">{{printf "%.2f" .UnitPrice}}</td>
        <td class="text-right">{{printf "%.2f" .Total}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>

  <div class="totals-container">
    <table class="totals-table">
      <tr>
        <td class="text-right" style="color: #64748b;">Subtotal:</td>
        <td class="text-right" style="font-weight: 600;">{{.Currency}} {{printf "%.2f" .Subtotal}}</td>
      </tr>
      <tr>
        <td class="text-right" style="color: #64748b;">Tax:</td>
        <td class="text-right" style="font-weight: 600;">{{.Currency}} {{printf "%.2f" .Tax}}</td>
      </tr>
      <tr class="grand-total">
        <td class="text-right">Total:</td>
        <td class="text-right">{{.Currency}} {{printf "%.2f" .TotalAmount}}</td>
      </tr>
    </table>
  </div>

  {{if .Notes}}
  <div style="margin-top: 24px; font-size: 12px; color: #64748b; background: #fafafa; padding: 10px; border-radius: 4px;">
    <strong>Notes:</strong> {{.Notes}}
  </div>
  {{end}}

  <div class="footer">
    Thank you for your business! This is a computer-generated document.
  </div>
</body>
</html>`
