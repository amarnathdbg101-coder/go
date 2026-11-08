# SPEC: go-contextkit

## 1. Metadata
- **Module**: `go-contextkit`
- **Package**: `gocontextkit`
- **Location**: `C:\Library\go-contextkit`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `context`, `net/http`, `crypto/rand`, `encoding/hex`, `strings`, `time`).

---

## 2. Core Capabilities
1. **Collision-Proof Type-Safe Context Keys**:
   - Stores Request ID, Correlation ID, User ID, Tenant ID, and Client IP using unexported typed keys, avoiding package collisions.
2. **Detached Context (`Detach`)**:
   - Decouples background goroutine execution from HTTP request cancellations while retaining all contextual values (`UserID`, `RequestID`).
   - Prevents premature cancellation of async audit logging or webhook workers.
3. **HTTP Tracing Middleware**:
   - Injects/echoes `X-Request-ID` and extracts real client IP behind reverse proxies (`X-Forwarded-For`, `X-Real-IP`).

---

## 3. Public API
```go
func WithRequestID(ctx context.Context, id string) context.Context
func GetRequestID(ctx context.Context) string

func WithCorrelationID(ctx context.Context, id string) context.Context
func GetCorrelationID(ctx context.Context) string

func WithUserID(ctx context.Context, userID string) context.Context
func GetUserID(ctx context.Context) string

func WithTenantID(ctx context.Context, tenantID string) context.Context
func GetTenantID(ctx context.Context) string

func WithClientIP(ctx context.Context, ip string) context.Context
func GetClientIP(ctx context.Context) string

func WithMetadata(ctx context.Context, key, val string) context.Context
func GetMetadata(ctx context.Context, key string) (string, bool)

func Detach(ctx context.Context) context.Context
func GenerateID() string
func HTTPMiddleware(next http.Handler) http.Handler
```
