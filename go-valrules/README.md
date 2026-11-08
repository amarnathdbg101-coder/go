# go-valrules 🛡️

> **Enterprise Format & Checksum Validation Suite for Go (Go 1.23+).**  
> Payment gateways, KYC systems, taxation software, aur telecom applications ke liye standard rules: Luhn Checksum (Cards/IMEI), Indian GSTIN, PAN, IFSC, E.164 Phones, MAC, UUID, Hex, aur URLs.

---

## 🚀 Features

- **Financial Validation**:
  - `ValidateLuhn`: Credit card aur debit card Mod-10 checksum validation.
  - `ValidateGSTIN`: 15-digit GSTIN validation + automatic State Name extraction (e.g. 27 -> Maharashtra).
  - `ValidatePAN`: Income tax PAN format validation.
  - `ValidateIFSC`: RBI banking IFSC code structure check.
- **Telecom & Identity**:
  - `ValidateE164`: International ITU-T phone numbers (`+91...`, `+1...`).
  - `ValidateIndianMobile`: 10-digit Indian numbers (`+91`, `0` prefixes handled).
  - `ValidateUUID`: RFC 4122 v4/v7 UUID verification.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-valrules"
)

func main() {
	// 1. Credit Card Checksum (Luhn Algorithm)
	isValidCard := valrules.ValidateLuhn("4242 4242 4242 4242")
	fmt.Printf("Card Valid: %v\n", isValidCard)

	// 2. GSTIN Validation & Parsing
	details, err := valrules.ValidateGSTIN("27ABCDE1234F1Z5")
	if err == nil {
		fmt.Printf("State: %s, PAN: %s\n", details.StateName, details.PAN)
	}

	// 3. Indian Mobile & IFSC
	fmt.Printf("Mobile: %v\n", valrules.ValidateIndianMobile("+91 9876543210"))
	fmt.Printf("IFSC: %v\n", valrules.ValidateIFSC("HDFC0000001"))
}
```
