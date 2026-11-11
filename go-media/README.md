# go-media 📄

> **Unified Documents, Media Processing & Fast PDF Suite for Go.**  
> Combines Sub-20ms Native PDF Generation, POS Thermal Receipts, Image Uploads, MIME Sniffing, and CSV Chunking into a single modular Go package.

---

## 📦 Sub-packages

- `github.com/shopsilo/go-media/gopdf`: Fast sub-20ms native Go PDF generator for A4 invoices & 80mm POS thermal slips
- `github.com/shopsilo/go-media/filekit`: 512B Magic byte MIME detection, image resizing, S3/Local storage
- `github.com/shopsilo/go-media/csvkit`: OOM-proof CSV streaming & Excel BOM stripper

---

## ⚡ 3-Minute Quickstart

```go
package main

import (
	"fmt"

	"github.com/shopsilo/go-media/gopdf"
)

func main() {
	receipt := gopdf.ReceiptData{
		ShopName:    "Gupta Kirana Store",
		BillNumber:  "BILL-101",
		Date:        "11 Nov 2026",
		TotalAmount: 350.0,
	}

	fmt.Printf("Receipt data ready for %s\n", receipt.ShopName)
}
```
