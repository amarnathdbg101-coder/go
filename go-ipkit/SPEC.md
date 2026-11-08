# SPEC: go-ipkit

## 1. Metadata
- **Module**: `go-ipkit`
- **Package**: `goipkit`
- **Location**: `C:\Library\go-ipkit`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `net`, `net/http`, `strings`, `fmt`).

---

## 2. Core Capabilities
1. **Trusted Proxy Backward Traversal (Anti-Spoofing)**:
   - Walks the `X-Forwarded-For` chain from right-to-left, identifying the true client without blindly trusting spoofed leftmost headers sent by malicious attackers.
2. **Reverse Proxy & CDN Header Extraction**:
   - Supports `CF-Connecting-IP` (Cloudflare) and `True-Client-IP` (Akamai).
3. **Subnet & CIDR Evaluation**:
   - `InCIDR(ip, subnet)`: Matches IPv4/IPv6 addresses against CIDR ranges.
   - `IsPrivate(ip)`: Detects RFC 1918 private subnets and loopback addresses.
4. **Port & Normalization**:
   - `CleanIP`: Strips IPv4/IPv6 port numbers and brackets.

---

## 3. Public API
```go
type Extractor struct { ... }
func New(opts ...Option) *Extractor
func (e *Extractor) GetRealIP(r *http.Request) string

func WithTrustedCIDRs(cidrs ...string) Option
func WithTrustPrivate(trust bool) Option

func CleanIP(addr string) string
func IsPrivate(ipStr string) bool
func InCIDR(ipStr, cidr string) (bool, error)
```
