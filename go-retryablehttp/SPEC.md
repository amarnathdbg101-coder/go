# SPEC: go-retryablehttp

## 1. Metadata
- **Module**: `go-retryablehttp`
- **Package**: `retryablehttp`
- **Location**: `C:\Library\go-retryablehttp`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `net/http`, `context`, `time`, `bytes`, `io`, `crypto/rand`, `strconv`).

---

## 2. Core Capabilities
1. **Automatic Request Body Rewinding**:
   - Stores payload streams into byte slices or invokes `req.GetBody()` across retries, solving the common Go pitfall where POST/PUT requests send empty bodies on subsequent attempts.
2. **Standard-Compliant `Retry-After` Header Handling**:
   - Parses both integer seconds (`Retry-After: 120`) and RFC 1123 HTTP-date format (`Retry-After: Wed, 21 Oct 2026 07:28:00 GMT`), pausing execution deterministically.
3. **Full-Jitter Exponential Backoff**:
   - Avoids thundering herd problem by applying randomized jitter within the computed exponential backoff interval.
4. **Context-Aware Abort & Connection Leak Protection**:
   - Immediately respects `ctx.Done()`, drains and closes stale response bodies to return idle sockets back to `http.Transport`.

---

## 3. Public API
```go
type RetryPolicy func(resp *http.Response, err error) bool
type OnRetryFunc func(attempt int, req *http.Request, resp *http.Response, err error)

type Config struct {
    MaxRetries  int
    MinWait     time.Duration
    MaxWait     time.Duration
    RetryPolicy RetryPolicy
    OnRetry     OnRetryFunc
    HTTPClient  *http.Client
}

type Client struct
func New(opts ...Option) *Client
func WithMaxRetries(retries int) Option
func WithWaitLimits(min, max time.Duration) Option
func WithRetryPolicy(policy RetryPolicy) Option
func WithOnRetry(fn OnRetryFunc) Option
func WithHTTPClient(httpClient *http.Client) Option

func NewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error)
func (c *Client) Do(req *http.Request) (*http.Response, error)
```
