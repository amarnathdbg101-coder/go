# go-pathkit 📁

> **Security-Hardened File Path & Atomic Filesystem Toolkit for Go (Go 1.23+).**  
> Path Traversal attacks (`../../etc/passwd`), atomic file writes (crash/power cut par zero data corruption), streaming SHA-256 file hashing, aur filtered directory walking ke liye production-ready utility suite.

---

## 🚀 Features

- **Path Traversal Defense**: `SafeJoin(baseDir, untrustedPath)` ensure karta hai ki user input base directory se bahar escape na kar sake.
- **Crash-Proof Atomic Write**: `WriteFileAtomic` temporary file me write karke `Sync()` karta hai aur atomically rename karta hai — process crash hone par bhi adhi-adhuri corrupt file nahi banti.
- **Streaming Hash**: `SHA256` aur `MD5` bina poori file ko RAM me load kiye streaming buffer se hash compute karte hain (multi-gigabyte files ke liye safe).
- **Directory Walking**: `Walk` me MaxDepth, Extensions whitelist (`.go`, `.json`), aur IgnoreDirs (`node_modules`, `.git`) support.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-pathkit"
	"log"
)

func main() {
	// 1. Safe Path Resolution (Traversal Guard)
	safePath, err := pathkit.SafeJoin("./uploads", "user_102/profile.png")
	if err != nil {
		log.Fatalf("Traversal attempt blocked: %v", err)
	}
	fmt.Println("Safe Path:", safePath)

	// 2. Atomic File Write
	configData := []byte(`{"port": 8080, "active": true}`)
	err = pathkit.WriteFileAtomic("./data/config.json", configData, 0644)
	if err != nil {
		log.Fatalf("Atomic write failed: %v", err)
	}
	fmt.Println("✅ File atomically written without corruption risk")
}
```
