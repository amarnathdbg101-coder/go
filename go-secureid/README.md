# go-secureid

Cryptographically secure random tokens, prefixed API keys, and NanoIDs for Go.

---

## Features
- **CSPRNG Powered**: Uses OS-level cryptographic random entropy (`crypto/rand`).
- **Zero Modulo Bias**: Bitmask filtering guarantees statistically uniform character distribution.
- **Prefixed API Keys**: Standard Stripe/GitHub-style credentials (`sk_live_...`, `pk_test_...`).
- **NanoID Compatible**: Compact, collision-resistant 21-character URL-safe identifiers.
- **Human-Readable Invite Codes**: Uses Crockford's Base32 alphabet (no confusing `0/O` or `1/I/L`).
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	gosecureid "go-secureid"
)

func main() {
	// 1. Generate Secret API Key
	apiKey, _ := gosecureid.APIKey("sk_live", 32)
	fmt.Println("API Key:     ", apiKey) // sk_live_k7h9_a...

	// 2. Generate NanoID for database primary key
	id, _ := gosecureid.NanoID(21)
	fmt.Println("NanoID:      ", id)

	// 3. Generate Human-Friendly Invite Code (e.g. for SMS / Referral)
	code, _ := gosecureid.InviteCode(8)
	fmt.Println("Invite Code: ", code) // e.g. 7K9M2V8P (No I, L, O, U)

	// 4. Generate URL-Safe CSRF Token
	token, _ := gosecureid.Token(32)
	fmt.Println("CSRF Token:  ", token)
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
