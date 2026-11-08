package pdfkit_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	pdfkit "github.com/shopsilo/go-chromedp-pdfkit"
)

// -------------------------------------------------------------------------
// 1. Template Rendering Unit Tests
// -------------------------------------------------------------------------

func TestTemplate_Render(t *testing.T) {
	tmpl := "<h1>Hello {{.Name}}! Balance: {{.Balance}}</h1>"
	data := map[string]any{"Name": "Alex", "Balance": 500}

	rendered, err := pdfkit.RenderHTML(tmpl, data)
	if err != nil {
		t.Fatalf("render template failed: %v", err)
	}

	expected := "<h1>Hello Alex! Balance: 500</h1>"
	if rendered != expected {
		t.Errorf("expected '%s', got '%s'", expected, rendered)
	}
}

// -------------------------------------------------------------------------
// 2. Headless Chrome PDF Generation Tests
// -------------------------------------------------------------------------

func TestGenerateFromHTML_ValidPDF(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	html := `<!DOCTYPE html>
<html>
<head><title>Test Doc</title></head>
<body style="font-family: sans-serif; padding: 20px;">
  <h1 style="color: #2563eb;">Antigravity PDFKit Test</h1>
  <p>Chromedp high-speed headless PDF rendering verified.</p>
</body>
</html>`

	opts := pdfkit.DefaultOptions()
	pdfBytes, err := pdfkit.GenerateFromHTML(ctx, html, opts)
	if err != nil {
		t.Fatalf("GenerateFromHTML failed: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatalf("expected non-empty PDF byte slice")
	}

	// Verify standard PDF magic header: starts with "%PDF-"
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("generated output is not a valid PDF document (missing %%PDF- header)")
	}
}

func TestGenerateInvoice_FullFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	invoice := pdfkit.InvoiceData{
		InvoiceNumber: "INV-2026-001",
		Date:          "07 Nov 2026",
		DueDate:       "14 Nov 2026",
		CompanyName:   "ShopSilo Technologies",
		CompanyEmail:  "billing@shopsilo.com",
		CompanyPhone:  "+91 98765 43210",
		CustomerName:  "Rajesh Sharma",
		CustomerEmail: "rajesh@client.com",
		Currency:      "₹",
		Items: []pdfkit.InvoiceItem{
			{Description: "Mechanical Gaming Keyboard", Quantity: 2, UnitPrice: 3499.00, Total: 6998.00},
			{Description: "Ergonomic Vertical Mouse", Quantity: 1, UnitPrice: 1899.00, Total: 1899.00},
		},
		Subtotal:    8897.00,
		Tax:         1601.46,
		TotalAmount: 10498.46,
		Notes:       "Warranty active for 1 year from invoice date.",
	}

	pdfBytes, err := pdfkit.GenerateInvoice(ctx, invoice)
	if err != nil {
		t.Fatalf("GenerateInvoice failed: %v", err)
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Fatalf("expected valid PDF header")
	}

	// Save to temp file to verify write helper
	tempDir, err := os.MkdirTemp("", "pdfkit_test_*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outPath := filepath.Join(tempDir, "invoice.pdf")
	if err := pdfkit.SaveToFile(pdfBytes, outPath); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	info, err := os.Stat(outPath)
	if err != nil || info.Size() == 0 {
		t.Errorf("saved PDF file is empty or missing")
	}
}

func TestGenerateThermalReceipt_FullFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	receipt := pdfkit.ReceiptData{
		ShopName:       "Gupta Kirana & Provisions",
		ShopAddress:    "Shop #12, Civil Lines, Prayagraj",
		ShopPhone:      "+91 98765 12345",
		GSTIN:          "09AAAAA0000A1Z5",
		BillNumber:     "BILL-1092",
		Date:           "07 Nov 2026, 08:30 PM",
		CustomerName:   "Amit Kumar",
		CustomerPhone:  "9876543210",
		Items: []pdfkit.ReceiptItem{
			{Name: "Fortune Sunlite Oil 1L", Qty: 2, Price: 145.0, Amount: 290.0},
			{Name: "Tata Salt 1kg", Qty: 1, Price: 28.0, Amount: 28.0},
			{Name: "Amul Butter 100g", Qty: 1, Price: 56.0, Amount: 56.0},
		},
		Subtotal:       374.0,
		TaxAmount:      18.7,
		TotalAmount:    392.7,
		PaymentMethod:  "UPI / Paytm",
		FooterNote:     "Thank you! Visit again for fresh groceries.",
	}

	pdfBytes, err := pdfkit.GenerateThermalReceipt(ctx, receipt, 80.0)
	if err != nil {
		t.Fatalf("GenerateThermalReceipt failed: %v", err)
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Fatalf("expected valid PDF header for thermal receipt")
	}
}

// -------------------------------------------------------------------------
// 3. Validation & Edge Cases Tests
// -------------------------------------------------------------------------

func TestEngine_ValidationErrors(t *testing.T) {
	ctx := context.Background()

	// Case 1: Empty HTML
	_, err := pdfkit.GenerateFromHTML(ctx, "", pdfkit.DefaultOptions())
	if err != pdfkit.ErrEmptyHTML {
		t.Errorf("expected ErrEmptyHTML, got %v", err)
	}

	// Case 2: Empty URL
	_, errURL := pdfkit.GenerateFromURL(ctx, "", pdfkit.DefaultOptions())
	if errURL != pdfkit.ErrEmptyURL {
		t.Errorf("expected ErrEmptyURL, got %v", errURL)
	}
}
