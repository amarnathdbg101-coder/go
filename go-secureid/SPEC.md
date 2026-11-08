# Technical Specification: go-secureid

## 1. Domain & Purpose
`go-secureid` is a cryptographically secure random token and identifier generator for Go. Powered by system CSPRNG (`crypto/rand`), it generates prefixed API credentials, collision-resistant NanoIDs, URL-safe session tokens, and human-friendly Crockford Base32 invitation codes with guaranteed zero modulo bias.

## 2. Architecture & Design Principles
- **CSPRNG Guaranteed**: Exclusively uses OS kernel entropy via `crypto/rand`.
- **Zero Modulo Bias**: Uses bitwise mask filtering over random byte buffers to prevent mathematical bias when mapping bytes into non-power-of-two alphabets.
- **Predefined Standard Alphabets**:
  - `NanoID`: URL-safe 64-character alphabet (`A-Z`, `a-z`, `0-9`, `_`, `-`).
  - `Crockford Base32`: Human-readable 32-character alphabet omitting visually ambiguous characters (`I`, `L`, `O`, `U`).
- **Zero Heavy Dependencies**: Pure Go standard library (`crypto/rand`, `encoding/base64`, `math/bits`, `strings`).

## 3. API Surface
- `func Token(byteLength int) (string, error)`
- `func APIKey(prefix string, byteLength int) (string, error)`
- `func NanoID(length int) (string, error)`
- `func InviteCode(length int) (string, error)`
- `func Custom(alphabet string, length int) (string, error)`

## 4. Error Handling
- `ErrInvalidLength`: Length parameter <= 0.
- `ErrEmptyAlphabet`: Alphabet length < 2.

## 5. Verification & Testing
- Unit tests verify Base64 URL safety, API key prefix delimiters (`sk_live_`), NanoID length and character sets, Crockford ambiguity exclusion, and 1,000-iteration collision-free generation.
