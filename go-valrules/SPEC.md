# SPEC: go-valrules

## 1. Metadata
- **Module**: `go-valrules`
- **Package**: `valrules`
- **Location**: `C:\Library\go-valrules`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `regexp`, `strconv`, `strings`, `net/url`, `unicode`, `encoding/base64`, `encoding/hex`).

---

## 2. Core Capabilities
1. **Financial & Checksum Validation**:
   - `ValidateLuhn(cardNumber)` verifies credit cards, debit cards, and IMEI identifiers via Mod-10.
   - `ValidateGSTIN(gstin)` verifies 15-character Indian GST numbers, extracts PAN, state code, and maps state names across all 36 Indian states & union territories.
   - `ValidatePAN(pan)` validates 10-character Indian Income Tax PAN card structure.
   - `ValidateIFSC(ifsc)` validates RBI Indian Financial System Codes.
2. **Telecom & Communication**:
   - `ValidateE164(phone)` ensures international ITU-T E.164 compliance (`+[1-9]\d{1,14}`).
   - `ValidateIndianMobile(mobile)` handles standard 10-digit, `+91`, and `0` prefixed numbers.
3. **Format & Data Integrity**:
   - `ValidateMAC(mac)`: Hexadecimal IEEE 802 MAC-48 address formats.
   - `ValidateUUID(uuid)`: RFC 4122 / RFC 9562 UUID strings.
   - `ValidateBase64(s)` & `ValidateHex(s)`: Low-overhead binary representation checks.
   - `ValidateURL(rawURL)`: Strict RFC URL validation with scheme whitelist and port bounds.

---

## 3. Public API
```go
var StateCodes map[string]string

type GSTINDetails struct {
    GSTIN     string
    StateCode string
    StateName string
    PAN       string
    EntityNum string
    Checksum  string
}

func ValidateLuhn(number string) bool
func ValidatePAN(pan string) bool
func ValidateGSTIN(gstin string) (GSTINDetails, error)
func ValidateIFSC(ifsc string) bool
func ValidateE164(phone string) bool
func ValidateIndianMobile(mobile string) bool
func ValidateMAC(mac string) bool
func ValidateUUID(uuidStr string) bool
func ValidateBase64(s string) bool
func ValidateHex(s string) bool
func ValidateURL(rawURL string) bool
```
