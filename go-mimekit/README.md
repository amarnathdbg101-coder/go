# go-mimekit 🔍

> **512-Byte Magic Byte File Sniffer & Polyglot Executable Defense for Go (Go 1.23+).**  
> File uploads me file extension (`.png`, `.jpg`) ko spoof karna aasan hota hai — attackers `.exe` ya shell script ka extension `.jpg` karke server par upload kar dete hain. `go-mimekit` pehle 512 bytes ke magic numbers ko inspect karta hai aur non-destructive stream rewinding deta hai.

---

## 🚀 Features

- **Genuine Magic Byte Detection**: PNG, JPEG, GIF, WebP, SVG, PDF, ZIP, GZIP, MP4, WAV auto-detected.
- **Executable Attack Defense**: Windows PE (`MZ`), Linux ELF, macOS Mach-O, aur Shell scripts (`#!`) ko identify karke block karta hai.
- **Stream Rewind (`SniffReader`)**: `io.Reader` se 512 bytes padhne ke baad bhi downstream handler (S3/Disk) ko poora stream 0th byte se wapas deta hai.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-mimekit"
	"os"
)

func main() {
	fileBytes, _ := os.ReadFile("untrusted_upload.bin")

	// 1. Sniff actual content
	ct := mimekit.Sniff(fileBytes)
	fmt.Printf("Detected MIME: %s (.%s)\n", ct.MIME, ct.Extension)

	// 2. Validate safe image upload
	if !mimekit.IsSafeImage(fileBytes) {
		fmt.Println("❌ Security Alert: File is not a valid safe image!")
		return
	}

	fmt.Println("✅ Safe image verified")
}
```
