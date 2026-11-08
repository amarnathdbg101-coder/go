# go-crypto 🔐

Lightweight, high-speed, zero-dependency cryptography kit for modern Go backends. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🛡️ **AES-256-GCM AEAD Encryption**: Military-grade authenticated encryption with 12-byte random nonce prepended. Bit-flipping attacks detect ho jate hain.
- 🔏 **HMAC-SHA256**: Constant-time signature verification for webhooks (Razorpay, Stripe, GitHub).
- 🆔 **UUIDv7 (RFC 9562)**: Millisecond time-ordered UUIDs for PostgreSQL primary keys (B-Tree friendly, prevents fragmentation).
- 🔢 **Cryptographic Random Tokens & OTP**: Unbiased 6-digit OTPs and URL-safe Base64 session tokens.
- 🔑 **PBKDF2-HMAC-SHA256**: Passphrase to 256-bit encryption key derivation.
- ⚡ **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. AES-256-GCM Encryption
```go
package main

import (
	"fmt"
	"log"
	gocrypto "go-crypto"
)

func main() {
	// 32-byte key for AES-256
	key, _ := gocrypto.RandomBytes(32)
	cipher, err := gocrypto.NewAES256(key)
	if err != nil {
		log.Fatal(err)
	}

	// Encrypt to URL-safe string
	secretData := "User Bank Account: 1234567890"
	ciphertext, err := cipher.EncryptString(secretData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Ciphertext:", ciphertext)

	// Decrypt back
	plaintext, err := cipher.DecryptString(ciphertext)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Plaintext:", plaintext)
}
```

### 2. Time-ordered UUIDv7 for Database Primary Keys
```go
id, err := gocrypto.NewUUIDv7()
if err != nil {
    log.Fatal(err)
}

// Example: 018e6931-92be-7b9c-8517-5735b54a7c1b
fmt.Println("New UUIDv7:", id.String())
```

### 3. Webhook HMAC Verification
```go
secret := []byte("webhook_secret")
payload := []byte(`{"event":"order_paid","amount":1500}`)

// Sign
sig := gocrypto.SignHMACSHA256Hex(payload, secret)

// Verify in Constant-Time (Side-channel resistant)
valid := gocrypto.VerifyHMACSHA256Hex(payload, sig, secret)
fmt.Println("Is signature valid?", valid)
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
