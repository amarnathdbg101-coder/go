# SPEC: go-gopdf-pdfkit

## 1. Scope & Purpose
Ultra-fast (sub-20ms) native PDF generation engine built in pure Go without headless browser dependencies.
- In Scope: Sub-20ms A4/A5 Tax Invoice rendering, sub-3ms 80mm/58mm POS thermal receipt generation with dynamic height, embedded Google Noto Sans TrueType fonts (`//go:embed`), and full UTF-8 Hindi + English + ₹ Rupee support.
- Out of Scope: Complex CSS3 flexbox/grid rendering (which is handled by `go-chromedp-pdfkit`).

## 2. Public API Contract
- `NewBuilder(format PageFormat, customHeight float64) (*PDFBuilder, error)`
- `GenerateInvoicePDF(inv Invoice, format PageFormat) ([]byte, error)`
- `GenerateReceiptPDF(rec Receipt, format PageFormat) ([]byte, error)`
- `TruncateRunes(s string, maxRunes int) string`: Clean multi-byte UTF-8 string truncation.

## 3. Concurrency & Memory Guarantees
- Memory footprint: Under 2 MB of heap per generation.
- Pure Go execution: No background threads, no external process spawns, instant garbage collection.

## 4. How to Extend
- Add new document layouts (e.g. shipping manifests, delivery challans, medical prescriptions) using the fluent `PDFBuilder` API.
