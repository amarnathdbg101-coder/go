# SPEC: go-cookie

## 1. Metadata
- **Module**: `go-cookie`
- **Package**: `gocookie`
- **Location**: `C:\Library\go-cookie`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `net/http`, `crypto/aes`, `crypto/cipher`, `crypto/rand`, `encoding/base64`, `errors`, `fmt`, `io`).

---

## 2. Core Capabilities
1. **Authenticated AES-256-GCM Encryption**:
   - Galois/Counter Mode authenticated encryption with unique 12-byte cryptographic nonces per cookie write.
   - Any client-side byte tampering or corruption is instantly detected and rejected (`ErrTamperedCookie`).
2. **Name-Bound Cryptographic Binding (Anti-Spoofing)**:
   - Uses the cookie's name as GCM Additional Authenticated Data (AAD), preventing attackers from copying an encrypted session cookie into a privileged admin cookie.
3. **Security by Default**:
   - Automatically enforces `HttpOnly: true` (XSS immune), `Secure: true` (HTTPS only), and `SameSite: Lax` (CSRF defense).

---

## 3. Public API
```go
type Jar struct { ... }
func New(secretKey32 []byte, opts ...Option) (*Jar, error)
func (j *Jar) Write(w http.ResponseWriter, cookie http.Cookie) error
func (j *Jar) Read(r *http.Request, name string) (string, error)
func (j *Jar) Delete(w http.ResponseWriter, name string)

func WithInsecureForDev() Option
func WithSameSite(mode http.SameSite) Option
```
