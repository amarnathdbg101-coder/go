# Technical Specification: go-webhook

## 1. Domain & Purpose
`go-webhook` is a secure, resilient outbound webhook delivery and inbound signature verification toolkit for Go. It implements standard cryptographic HMAC-SHA256 signatures with timestamp anti-replay verification (identical to industry specifications from Stripe and GitHub), concurrent worker-pooled delivery dispatching, exponential backoff retries on transient errors, and panic-free HTTP lifecycle management.

## 2. Architecture & Design Principles
- **Cryptographic Security**: Uses `crypto/hmac` and `crypto/sha256` with strict constant-time byte comparisons (`hmac.Equal`) to defend against timing attacks.
- **Anti-Replay Defense**: Signatures bind the Unix epoch timestamp `t=` into the MAC digest. Verification rejects stale requests older than the configured drift tolerance.
- **Resilient Delivery**: Dispatcher automatically retries transient server errors (HTTP 5xx and 429) using exponential backoff while immediately terminating on permanent 4xx client errors (400, 401, 404).
- **Asynchronous Worker Pool**: Provides bounded worker goroutines for high-throughput background dispatching without unbounded goroutine proliferation.

## 3. API Surface
- `func Sign(secret []byte, payload []byte, timestamp int64) (string, error)`
- `func Verify(secret []byte, payload []byte, signatureHeader string, tolerance time.Duration) error`
- `type DeliveryResult struct { StatusCode, Attempts int; Duration time.Duration; ResponseBody []byte; Success bool; Error error }`
- `type Config struct { Timeout, BackoffBase time.Duration; MaxRetries, WorkerPoolSize int; SignatureHeader, EventHeader, IDHeader, UserAgent string; HTTPClient *http.Client }`
- `func DefaultConfig() Config`
- `func NewDispatcher(cfg Config) *Dispatcher`
- `func (d *Dispatcher) Dispatch(ctx context.Context, targetURL, event, webhookID string, payload []byte, secret []byte) (*DeliveryResult, error)`
- `func (d *Dispatcher) DispatchAsync(ctx context.Context, targetURL, event, webhookID string, payload []byte, secret []byte, onComplete func(*DeliveryResult))`
- `func (d *Dispatcher) Stop()`

## 4. Error Handling
- `ErrInvalidSignature`: Cryptographic HMAC mismatch or signature corruption.
- `ErrSignatureExpired`: Timestamp age exceeds allowable verification drift tolerance.
- `ErrMalformedHeader`: Missing `t=` or `v1=` components in header.
- `ErrEmptySecret`: Secret key not supplied.
- `ErrMaxRetriesExceeded`: Delivery attempts exhausted without a 2xx response.

## 5. Verification & Testing
- Unit tests verify signature creation, constant-time validation, tampered payload rejection, expired timestamp rejection, HTTP delivery retry loops, non-retryable 4xx handling, and asynchronous worker pool dispatch.
