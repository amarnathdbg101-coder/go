# SPEC: go-media

## 1. Scope & Purpose
Unified documents, media processing, PDF generation, and file storage suite for Go applications.
Consolidates Sub-20ms Native PDF Generator, POS Thermal Receipts, Headless Chrome PDF, Image Resizing & MIME Sniffing, CSV Streaming, and Gzip Compression under one clean, modular package.

## 2. Public API Sub-packages
- `github.com/shopsilo/go-media/gopdf`: Fast sub-20ms native Go PDF generator for A4 invoices & 80mm POS thermal slips
- `github.com/shopsilo/go-media/filekit`: 512B Magic byte MIME detection, image resizing, S3/Local storage
- `github.com/shopsilo/go-media/csvkit`: OOM-proof CSV streaming & Excel BOM stripper

## 3. Concurrency & Memory Guarantees
- Sub-20ms native PDF generation. Zero Chrome dependency for gopdf.
- Path traversal defense & MIME spoofing protection.
