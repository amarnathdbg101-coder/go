# SPEC: go-crypto

## 1. Metadata
- **Module**: `go-crypto`
- **Package**: `gocrypto`
- **Location**: `C:\Library\go-crypto`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `crypto/aes`, `crypto/cipher`, `crypto/rand`, `crypto/hmac`, `crypto/sha256`, `crypto/subtle`, `encoding/hex`, `encoding/base64`).

---

## 2. Core Capabilities
1. **AES-256-GCM Authenticated Encryption**:
   - Galois/Counter Mode (AEAD) with 12-byte cryptographic nonce prepended.
   - Built-in support for Additional Authenticated Data (AAD) to prevent replay/tampering.
   - String encryption/decryption with URL-safe Base64 output.
2. **HMAC-SHA256 Signatures**:
   - Constant-time verification (`subtle.ConstantTimeCompare`) preventing timing attacks.
   - Raw byte and Hex signatures for webhooks (Stripe, Razorpay, GitHub).
3. **Cryptographically Secure Tokens**:
   - Unbiased N-digit numeric OTP generation (`big.Int` calculation).
   - Random Hex and URL-safe Base64 token generation.
4. **UUIDv7 (RFC 9562)**:
   - 48-bit millisecond timestamp + 74 bits of entropy.
   - Chronologically sortable for PostgreSQL B-tree primary key performance.
5. **PBKDF2-HMAC-SHA256 Key Derivation**:
   - RFC 8018 compliant key stretching from user passphrases.

---

## 3. Public API
```go
// AES-256-GCM
func NewAES(key []byte) (*AESCipher, error)
func NewAES256(key []byte) (*AESCipher, error)
func (c *AESCipher) Encrypt(plaintext []byte) ([]byte, error)
func (c *AESCipher) Decrypt(ciphertextWithNonce []byte) ([]byte, error)
func (c *AESCipher) EncryptWithAAD(plaintext, aad []byte) ([]byte, error)
func (c *AESCipher) DecryptWithAAD(ciphertextWithNonce, aad []byte) ([]byte, error)
func (c *AESCipher) EncryptString(plaintext string) (string, error)
func (c *AESCipher) DecryptString(base64Ciphertext string) (string, error)

// HMAC-SHA256
func SignHMACSHA256(message, secret []byte) []byte
func SignHMACSHA256Hex(message, secret []byte) string
func VerifyHMACSHA256(message, signature, secret []byte) bool
func VerifyHMACSHA256Hex(message []byte, hexSignature string, secret []byte) bool

// Tokens & Hashes
func RandomBytes(n int) ([]byte, error)
func RandomHex(bytesLen int) (string, error)
func RandomBase64URL(bytesLen int) (string, error)
func RandomNumericOTP(digits int) (string, error)
func HashSHA256(data []byte) []byte
func HashSHA256Hex(data []byte) string

// UUIDv7
func NewUUIDv7() (UUID, error)
func (u UUID) String() string
func ParseUUID(s string) (UUID, error)

// PBKDF2
func DeriveKeyPBKDF2(passphrase string, salt []byte, iterations, keyLen int) ([]byte, error)
```
