# go-checksum

Streaming file integrity checksum calculation and verification engine for Go.

---

## Features
- **Multi-Algorithm Support**: SHA-256, SHA-512, MD5, SHA-1, CRC32 (IEEE), and Adler-32.
- **Constant Memory Footprint**: Uses `sync.Pool` 32KB buffers; can hash 100GB files in under 1MB of memory.
- **Timing Attack Safe**: Validated using constant-time `hmac.Equal`.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	gochecksum "go-checksum"
)

func main() {
	// 1. Calculate SHA-256 for a file
	hash, err := gochecksum.FromFile("release_v1.0.zip", gochecksum.SHA256)
	if err != nil {
		panic(err)
	}
	fmt.Println("SHA-256:", hash)

	// 2. Verify file integrity
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	valid, _ := gochecksum.VerifyFile("release_v1.0.zip", expected, gochecksum.SHA256)
	if valid {
		fmt.Println("File verification PASSED!")
	}
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
