# go-image-filekit 🖼️

> **Production-Grade, Zero-Cgo Image & File Upload Engine for Go.**  
> Features Pure Go Image Resizing, Magic Bytes Spoofing Defense, Path Traversal Protection, and Pluggable Storage Drivers.

---

## ✨ Features

- 🏎️ **Zero Cgo / Zero External Dependencies**: 100% Pure Go Standard Library (`image`, `image/jpeg`, `image/png`, `net/http`, `os`). Compiles instantly on Windows, Linux, and macOS without installing `libvips` or `imagemagick`.
- 🛡️ **Anti-Spoofing (Magic Bytes Sniffing)**: Reads real byte signatures via `http.DetectContentType` to permanently block attackers from uploading `malware.exe` renamed to `avatar.png`.
- 🔒 **Path Traversal Protection**: Enforces strict boundary checks to prevent `../../etc/passwd` directory escaping attacks.
- 📐 **Pure Go Aspect-Ratio Resizing**: High-speed bilinear downscaling (e.g. max `1200x1200` or `300x300` thumbnails) preserving natural aspect ratio without distortion.
- 🗜️ **JPEG Compression**: Automated quality tuning (default 85%) saving 60-70% storage and bandwidth.
- 🗄️ **Pluggable Storage (`Storage` Interface)**:
  - `LocalStorage`: Production local filesystem storage with auto-directory creation and public URL generation.
  - S3 / Cloudflare R2 / MinIO compatible interface.
- 🌐 **Multipart HTTP Helper**: One-line form file upload directly from `*http.Request`.

---

## 📦 Installation

```bash
go get github.com/amarnathdbg101-coder/go/go-image-filekit
```

---

## ⚡ 2-Minute Quickstart

### In Chi / net/http Handler (`POST /upload/product-image`)

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	filekit "github.com/amarnathdbg101-coder/go/go-image-filekit"
)

func main() {
	// 1. Initialize Local Storage (e.g. saves to ./uploads, serves at /static/uploads)
	storage, err := filekit.NewLocalStorage("./uploads", "http://localhost:8080/static/uploads")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Configure Image Uploader with auto-resize and compression
	uploader, err := filekit.NewUploader(storage,
		filekit.WithSubDir("products"),             // Saves in ./uploads/products/
		filekit.WithMaxFileSize(5*1024*1024),       // Max 5 MB
		filekit.WithImageResize(1200, 1200),        // Downscales large photos to fit 1200x1200
		filekit.WithJPEGQuality(80),                // 80% compression quality
	)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	// 3. File Upload HTTP Endpoint
	r.Post("/upload", func(w http.ResponseWriter, req *http.Request) {
		// Automatically reads multipart form field "image", verifies magic bytes,
		// resizes, compresses, saves to disk, and returns metadata!
		result, err := uploader.UploadFromRequest(req, "image")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// Serve uploaded files statically
	r.Handle("/static/uploads/*", http.StripPrefix("/static/uploads/", http.FileServer(http.Dir("./uploads"))))

	log.Println("Server running at :8080")
	http.ListenAndServe(":8080", r)
}
```

#### Sample JSON Response:

```json
{
  "filename": "1731003450123456789_a1b2c3d4e5f60718.jpg",
  "url": "http://localhost:8080/static/uploads/products/1731003450123456789_a1b2c3d4e5f60718.jpg",
  "size": 184520,
  "mime_type": "image/jpeg",
  "width": 1200,
  "height": 900,
  "original_name": "gadget_photo.png"
}
```

---

## 🛡️ Security Architecture

### 1. Magic Bytes Sniffing (MIME Verification)
```
Incoming Upload: "profile.png"
       │
       ▼
Read first 512 bytes ──► Check real magic signature
       │
       ├── Starts with "\x89PNG\r\n\x1a\n" ──► ✅ Legitimate PNG -> Process
       └── Starts with "#!/bin/bash" or "MZ" ──► ❌ Blocked! ErrInvalidMimeType
```

### 2. Collision-Free Random Filenames
Filename format:
```
<unix_nano_timestamp>_<8_random_crypto_hex_bytes>.<verified_ext>
Example: 1731003450123456789_f3a9c8b7e2d1045a.jpg
```
- Eliminates file overwrite collisions under high concurrency.
- Prevents attackers from uploading executable extensions like `.php`, `.jsp`, or `.sh`.

---

## 🧪 Testing

```bash
cd C:\Library\go-image-filekit
go test -v -count=1 ./...
```
