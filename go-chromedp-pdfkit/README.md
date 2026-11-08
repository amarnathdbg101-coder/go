# go-chromedp-pdfkit 🖨️

> **High-Speed, Anti-Zombie Headless Chrome PDF Generation Engine for Go.**  
> Built with `chromedp`, zero memory leaks, modern CSS3/HTML5 rendering, and built-in Invoice/Barcode templates.

---

## ✨ Features

- ⚡ **Pixel-Perfect Rendering**: Uses Google Chrome / Microsoft Edge rendering engine for 100% accurate CSS Flexbox, Grid, Custom Fonts, Barcodes, and SVGs.
- 🛡️ **Anti-Zombie Process Protection**: 3-Tier context hierarchy guarantees that every `chrome.exe` process is terminated immediately upon completion or timeout (zero memory leaks).
- 📜 **Versatile Input Sources**:
  - `GenerateFromHTML(ctx, htmlString, opts)`: Base64 data URL rendering for dynamic HTML.
  - `GenerateFromTemplate(ctx, tmplString, data, opts)`: Safe Go `html/template` evaluation.
  - `GenerateFromURL(ctx, targetURL, opts)`: Live webpage rendering.
  - `GenerateInvoice(ctx, invoiceData)`: Ready-to-use A4 retail & B2B invoice generation in **1 line of code**.
- 📐 **Flexible Paper Layouts**: A4, US Letter, Legal, and custom Thermal Barcode Label sizes (e.g. 4x6" shipping labels).
- 🎨 **Print Background Graphics**: Preserves CSS colors, badges, table striping, and gradient backgrounds.
- ⏱️ **Wait Triggers**: Wait for specific CSS selectors (`#invoice-ready`) or custom sleep durations for web fonts.

---

## 📦 Installation

```bash
go get github.com/shopsilo/go-chromedp-pdfkit
```

---

## ⚡ 2-Minute Quickstart

### 1. Generate Invoice PDF in 1 Line

```go
package main

import (
	"context"
	"log"

	pdfkit "github.com/shopsilo/go-chromedp-pdfkit"
)

func main() {
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
		Notes:       "Warranty valid for 1 year from invoice date.",
	}

	// 1-line generation using built-in high-quality CSS template
	pdfBytes, err := pdfkit.GenerateInvoice(context.Background(), invoice)
	if err != nil {
		log.Fatalf("PDF generation failed: %v", err)
	}

	// Save to file
	if err := pdfkit.SaveToFile(pdfBytes, "invoice.pdf"); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Invoice generated successfully: invoice.pdf")
}
```

---

## 🌐 Chi / net/http API Route Demo (`GET /invoices/{id}/download`)

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	pdfkit "github.com/shopsilo/go-chromedp-pdfkit"
)

func handleDownloadInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID := chi.URLParam(r, "id")

	// 1. Fetch invoice data from database (PostgreSQL)
	invoiceData := fetchInvoiceFromDB(invoiceID)

	// 2. Generate PDF bytes
	pdfBytes, err := pdfkit.GenerateInvoice(r.Context(), invoiceData)
	if err != nil {
		http.Error(w, "Failed to render invoice PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Stream PDF buffer with proper headers for browser preview/download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"invoice_%s.pdf\"", invoiceID))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}
```

---

## 🛡️ Anti-Zombie Chrome Process Architecture

Raw `chromedp` implementations often leave orphaned `chrome.exe` processes running when requests are cancelled or time out, eventually crashing the server.

`go-chromedp-pdfkit` enforces a **3-tier guaranteed context tree**:

```
[Request / Root Context]
       │
       ▼
1. ExecAllocator Context  ──► defer cancelAlloc() (KILLS chrome.exe process on exit)
       │
       ▼
2. Tab Context            ──► defer cancelTask() (CLOSES browser tab)
       │
       ▼
3. Timeout Context        ──► defer cancelTimeout() (ABORTS on 30s deadline)
```

Even if a client drops their connection or network lags, all Chromium processes and memory are cleanly freed within milliseconds.

---

## 📐 Supported Layout Options

```go
// 1. Standard A4
opts := pdfkit.DefaultOptions()

// 2. Retail Invoice (A4, 0.3in margins, background graphics enabled)
opts := pdfkit.InvoiceOptions()

// 3. Thermal Barcode / Shipping Label (e.g. 4x6 inches, 0 margins)
opts := pdfkit.BarcodeLabelOptions(4.0, 6.0)

// 4. Custom Landscape with Header/Footer
opts := pdfkit.PDFOptions{
    PaperWidth:          8.27,
    PaperHeight:         11.69,
    Landscape:           true,
    PrintBackground:     true,
    DisplayHeaderFooter: true,
    FooterTemplate:      `<div style="font-size:10px;text-align:center;width:100%;">Page <span class="pageNumber"></span> of <span class="totalPages"></span></div>`,
}
```

---

## 🧪 Testing

```bash
cd C:\Library\go-chromedp-pdfkit
go test -v -count=1 ./...
```
