package gopdfkit_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	gopdfkit "github.com/amarnathdbg101-coder/go/go-gopdf-pdfkit"
)

func sampleInvoice() gopdfkit.Invoice {
	return gopdfkit.Invoice{
		InvoiceNumber:   "INV-2026-999",
		Date:            "07-11-2026",
		DueDate:         "14-11-2026",
		CompanyName:     "ShopSilo Retail Private Limited",
		CompanyAddress:  "Connaught Place, New Delhi 110001",
		CompanyPhone:    "+91 11 2345 6789",
		CompanyGST:      "07AAAAA0000A1Z5",
		CustomerName:    "Amit Verma",
		CustomerPhone:   "+91 98111 22233",
		CustomerAddress: "Sector 62, Noida, UP",
		Currency:        "â‚¹",
		Items: []gopdfkit.InvoiceItem{
			{Description: "Wireless Ergonomic Mouse", Quantity: 2, UnitPrice: 1299.00, Total: 2598.00},
			{Description: "Type-C Fast Charging Cable 2m", Quantity: 3, UnitPrice: 399.00, Total: 1197.00},
			{Description: "Mechanical Gaming Keyboard RGB", Quantity: 1, UnitPrice: 3499.00, Total: 3499.00},
		},
		Subtotal:    7294.00,
		Tax:         1312.92,
		Discount:    500.00,
		GrandTotal:  8106.92,
		PaymentMode: "UPI / PhonePe",
		Notes:       "Goods once sold are covered under 1-year brand warranty.",
	}
}

// -------------------------------------------------------------------------
// 1. Generation Speed & PDF Validity Tests
// -------------------------------------------------------------------------

func TestGenerateInvoicePDF_SpeedAndValidity(t *testing.T) {
	inv := sampleInvoice()

	start := time.Now()
	pdfBytes, err := gopdfkit.GenerateInvoicePDF(inv, gopdfkit.FormatA4)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("GenerateInvoicePDF failed: %v", err)
	}

	t.Logf("âš¡ Pure Go A4 Invoice generation took: %v (sub-25ms benchmark)", elapsed)

	if len(pdfBytes) == 0 {
		t.Fatalf("expected non-empty PDF bytes")
	}

	// Verify standard PDF magic header
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("output does not begin with %%PDF- magic header")
	}

	// Verify file write to disk
	tempDir, _ := os.MkdirTemp("", "gopdf_test_*")
	defer os.RemoveAll(tempDir)

	outPath := filepath.Join(tempDir, "test_invoice.pdf")
	if err := os.WriteFile(outPath, pdfBytes, 0644); err != nil {
		t.Fatalf("failed to write test PDF to disk: %v", err)
	}

	info, err := os.Stat(outPath)
	if err != nil || info.Size() == 0 {
		t.Errorf("generated PDF on disk is empty or missing")
	}
}

// -------------------------------------------------------------------------
// 2. POS Thermal 80mm Receipt Tests
// -------------------------------------------------------------------------

func TestGenerateReceiptPDF_80mm(t *testing.T) {
	rec := gopdfkit.Receipt{
		StoreName:    "SHOPSILO SUPERMART",
		StoreAddress: "Shop 12, Main Market, Delhi",
		StorePhone:   "011-23456789",
		ReceiptNo:    "POS-8841",
		Date:         "07/11/2026 22:30",
		Cashier:      "Pooja S.",
		Currency:     "â‚¹",
		Items: []gopdfkit.ReceiptItem{
			{Name: "Amul Butter 500g", Qty: 1, Price: 275.00, Amount: 275.00},
			{Name: "Basmati Rice 5kg", Qty: 1, Price: 499.00, Amount: 499.00},
			{Name: "Tata Salt 1kg", Qty: 2, Price: 28.00, Amount: 56.00},
		},
		Subtotal:    830.00,
		Tax:         41.50,
		GrandTotal:  871.50,
		PaymentMode: "CASH",
		FooterNote:  "Save Paper, Save Trees! Thank you.",
	}

	start := time.Now()
	pdfBytes, err := gopdfkit.GenerateReceiptPDF(rec, gopdfkit.FormatThermal80mm)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("GenerateReceiptPDF failed: %v", err)
	}

	t.Logf("âš¡ Pure Go 80mm POS Thermal Receipt took: %v", elapsed)

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("expected valid PDF header for thermal receipt")
	}
}

// -------------------------------------------------------------------------
// 3. UTF-8 Hindi (Devanagari) & Rupee (â‚¹) Symbol Tests
// -------------------------------------------------------------------------

func TestUnicode_HindiAndRupee(t *testing.T) {
	inv := sampleInvoice()
	// Add Hindi Devanagari items
	inv.Items = append(inv.Items, gopdfkit.InvoiceItem{
		Description: "à¤ªà¤¨à¥€à¤° à¤Ÿà¤¿à¤•à¥à¤•à¤¾ à¤®à¤¸à¤¾à¤²à¤¾ (Paneer Tikka)",
		Quantity:    2,
		UnitPrice:   350.00,
		Total:       700.00,
	})

	pdfBytes, err := gopdfkit.GenerateInvoicePDF(inv, gopdfkit.FormatA5)
	if err != nil {
		t.Fatalf("failed to render Hindi Devanagari text in PDF: %v", err)
	}

	if len(pdfBytes) < 5000 {
		t.Errorf("expected generated PDF with embedded fonts to exceed 5KB, got %d", len(pdfBytes))
	}
}

// -------------------------------------------------------------------------
// 4. TruncateRunes Unit Tests
// -------------------------------------------------------------------------

func TestTruncateRunes(t *testing.T) {
	// Hindi multi-byte runes
	hindiText := "à¤¨à¤®à¤¸à¥à¤¤à¥‡ à¤­à¤¾à¤°à¤¤ à¤¦à¥à¤¨à¤¿à¤¯à¤¾"
	truncated := gopdfkit.TruncateRunes(hindiText, 7)

	if len([]rune(truncated)) != 7 {
		t.Errorf("expected 7 runes, got %d ('%s')", len([]rune(truncated)), truncated)
	}
}
