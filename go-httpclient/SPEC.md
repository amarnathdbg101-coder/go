# SPEC: go-httpclient

## 1. Metadata
- **Module**: `go-httpclient`
- **Package**: `gohttpclient`
- **Location**: `C:\Library\go-httpclient`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `net/http`, `net/url`, `context`, `encoding/json`, `time`, `sync`, `math/rand`, `errors`, `fmt`, `io`, `bytes`).

---

## 2. Core Capabilities
1. **Resilient Connection Pooling**:
   - Tuned `http.Transport` with Keep-Alive, idle connection limits, TLS handshake timeouts, and header timeout enforcement.
2. **Automatic Exponential Backoff with Full Jitter**:
   - Retries on network disconnects, timeouts, or transient HTTP 5xx/429 errors.
   - Full jitter prevents synchronized thundering herd spikes.
   - Context cancellation checked between retry attempts.
3. **Built-in Circuit Breaker**:
   - `Closed`, `Open`, `HalfOpen` states prevent cascading microservice outages.
   - Fast fails without network requests when open.
4. **Leak-Proof Connection Management**:
   - Guarantee of draining body to EOF (`io.Copy(io.Discard, resp.Body)`) and closing body to prevent socket leakage.
5. **Fluent JSON APIs**:
   - Auto-marshaling of request payload and auto-unmarshaling of response payload directly into target pointers.

---

## 3. Public API
```go
type Client struct { ... }

type Option func(*Client)
func WithTimeout(d time.Duration) Option
func WithRetry(cfg RetryConfig) Option
func WithCircuitBreaker(maxFailures int, cooldown time.Duration) Option
func WithHeader(key, value string) Option

func New(opts ...Option) *Client

func (c *Client) Get(ctx context.Context, urlStr string, target any) (*http.Response, error)
func (c *Client) Post(ctx context.Context, urlStr string, body any, target any) (*http.Response, error)
func (c *Client) Put(ctx context.Context, urlStr string, body any, target any) (*http.Response, error)
func (c *Client) Delete(ctx context.Context, urlStr string, target any) (*http.Response, error)
func (c *Client) Do(req *http.Request, target any) (*http.Response, error)

func AddQueryParams(baseURL string, params map[string]string) (string, error)
```
