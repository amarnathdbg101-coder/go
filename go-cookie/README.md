# go-cookie 🍪

Cryptographically encrypted and authenticated HTTP cookies using AES-256-GCM with strict HttpOnly, Secure, and SameSite defenses. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🛡️ **AES-256-GCM AEAD Encryption**: Cookie values browser me completely encrypted Base64URL string ban kar jaati hain.
- 🚯 **Anti-Tampering & Anti-Spoofing**: Cookie ka name cryptographic Additional Authenticated Data (AAD) me bind hota hai; agar attacker payload ko rename ya alter karega to decryption reject ho jayegi.
- 🔒 **Security by Default**: Har cookie me `HttpOnly: true`, `Secure: true`, `SameSite: Lax`, aur `Path: "/"` automatically set hota hai.
- 📦 **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Setting and Reading Encrypted Cookies
```go
package main

import (
	"crypto/rand"
	"net/http"
	gocookie "go-cookie"
)

func main() {
	// 32-byte master encryption key
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	jar, _ := gocookie.New(key)

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// Encrypt and write session token
		_ = jar.Write(w, http.Cookie{
			Name:  "session_id",
			Value: "user_secret_token_1029",
		})
		w.Write([]byte("Logged in!"))
	})

	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		// Decrypt and authenticate session token
		token, err := jar.Read(r, "session_id")
		if err != nil {
			http.Error(w, "Unauthorized or tampered cookie", http.StatusUnauthorized)
			return
		}
		w.Write([]byte("Welcome! Token: " + token))
	})

	http.ListenAndServe(":8080", nil)
}
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
