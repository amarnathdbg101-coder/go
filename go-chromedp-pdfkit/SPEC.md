# SPEC: go-chromedp-pdfkit

## 1. Scope & Purpose
Headless Chrome / Edge browser automation engine for generating pixel-perfect PDF documents from HTML, CSS3, web URLs, and Go templates.
- In Scope: 3-Tier Anti-Zombie context guarantee, HTML to PDF, URL to PDF, Go `html/template` rendering, A4/Letter/Thermal Barcode label sizing, and CSS background graphics printing.
- Out of Scope: General web scraping or crawler scheduling.

## 2. Public API Contract
- `New(opts ...EngineOption) *Engine`
- `Engine.GenerateFromHTML(ctx, html, opts) ([]byte, error)`
- `Engine.GenerateFromURL(ctx, url, opts) ([]byte, error)`
- `Engine.GenerateFromTemplate(ctx, tmpl, data, opts) ([]byte, error)`
- `GenerateInvoice(ctx, invoiceData) ([]byte, error)`: 1-line built-in modern A4 invoice.
- `DefaultOptions()`, `InvoiceOptions()`, `BarcodeLabelOptions(w, h)`: Layout presets.

## 3. Concurrency & Memory Guarantees
- Anti-Zombie Process Guarantee: Allocator context, tab context, and execution timeout context are strictly wrapped in `defer cancel()` blocks. Orphaned Chromium processes are strictly prevented.

## 4. How to Extend
- Add custom browser flags via `WithAdditionalFlags` or specify custom browser binaries via `WithExecPath`.
