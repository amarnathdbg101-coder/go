package retryablehttp

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RetryPolicy determines whether an HTTP request should be retried based on response or error.
type RetryPolicy func(resp *http.Response, err error) bool

// OnRetryFunc is a callback invoked before a retry sleep occurs.
type OnRetryFunc func(attempt int, req *http.Request, resp *http.Response, err error)

// Config holds resilient client tuning parameters.
type Config struct {
	MaxRetries  int
	MinWait     time.Duration
	MaxWait     time.Duration
	RetryPolicy RetryPolicy
	OnRetry     OnRetryFunc
	HTTPClient  *http.Client
}

// DefaultConfig provides production-ready defaults.
func DefaultConfig() Config {
	return Config{
		MaxRetries:  3,
		MinWait:     200 * time.Millisecond,
		MaxWait:     2 * time.Second,
		RetryPolicy: DefaultRetryPolicy,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DefaultRetryPolicy retries on network errors, 429 Too Many Requests, and 5xx Server Errors.
func DefaultRetryPolicy(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp == nil {
		return false
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
		return true
	}
	return false
}

// Client wraps standard *http.Client with automatic retries and payload rewinding.
type Client struct {
	cfg Config
	mu  sync.RWMutex
}

// New creates a new Client with given options.
func New(opts ...Option) *Client {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Client{cfg: cfg}
}

// Option configures the Client.
type Option func(*Config)

// WithMaxRetries sets the maximum retry attempts.
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		if retries >= 0 {
			c.MaxRetries = retries
		}
	}
}

// WithWaitLimits sets min and max backoff boundaries.
func WithWaitLimits(min, max time.Duration) Option {
	return func(c *Config) {
		if min > 0 {
			c.MinWait = min
		}
		if max >= min {
			c.MaxWait = max
		}
	}
}

// WithRetryPolicy sets a custom retry decision policy.
func WithRetryPolicy(policy RetryPolicy) Option {
	return func(c *Config) {
		if policy != nil {
			c.RetryPolicy = policy
		}
	}
}

// WithOnRetry registers an observability callback for retries.
func WithOnRetry(fn OnRetryFunc) Option {
	return func(c *Config) {
		c.OnRetry = fn
	}
}

// WithHTTPClient overrides the underlying standard *http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Config) {
		if httpClient != nil {
			c.HTTPClient = httpClient
		}
	}
}

// NewRequest creates an http.Request with rewindable GetBody from bytes, string, or reader.
func NewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	if bodyBytes != nil {
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
		req.ContentLength = int64(len(bodyBytes))
	}

	return req, nil
}

// Do executes an HTTP request with automatic retries and exponential jitter backoff.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// If body exists but GetBody is nil, attempt to capture it
	if req.Body != nil && req.GetBody == nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		// Ensure context is not canceled
		if errCtx := req.Context().Err(); errCtx != nil {
			if resp != nil {
				_ = resp.Body.Close()
			}
			return nil, errCtx
		}

		// Rewind body if retrying
		if attempt > 0 && req.GetBody != nil {
			body, errBody := req.GetBody()
			if errBody != nil {
				if resp != nil {
					_ = resp.Body.Close()
				}
				return nil, errBody
			}
			req.Body = body
		}

		resp, err = c.cfg.HTTPClient.Do(req)

		// Check if we should retry
		shouldRetry := c.cfg.RetryPolicy(resp, err)
		if !shouldRetry {
			return resp, err
		}

		// If this was our last attempt, break
		if attempt == c.cfg.MaxRetries {
			break
		}

		// Notify callback if configured
		if c.cfg.OnRetry != nil {
			c.cfg.OnRetry(attempt+1, req, resp, err)
		}

		// Calculate sleep time, taking Retry-After into account
		wait := c.calculateBackoff(attempt, resp)

		// Drain and close previous body before sleeping to prevent connection leaks
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		// Sleep or abort on context cancellation
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(wait):
		}
	}

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// calculateBackoff computes exponential backoff with full jitter and parses Retry-After header.
func (c *Client) calculateBackoff(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
			if seconds, err := strconv.Atoi(retryAfterStr); err == nil && seconds > 0 {
				return time.Duration(seconds) * time.Second
			}
			if targetTime, err := http.ParseTime(retryAfterStr); err == nil {
				if wait := time.Until(targetTime); wait > 0 {
					return wait
				}
			}
		}
	}

	// Exponential backoff: minWait * 2^attempt
	multiplier := 1 << attempt
	backoff := c.cfg.MinWait * time.Duration(multiplier)
	if backoff > c.cfg.MaxWait || backoff <= 0 {
		backoff = c.cfg.MaxWait
	}

	// Full jitter: random duration between 0 and backoff
	var b [8]byte
	_, _ = rand.Read(b[:])
	r := binary.LittleEndian.Uint64(b[:])
	jitter := time.Duration(r % uint64(backoff))

	if jitter < c.cfg.MinWait {
		return c.cfg.MinWait
	}
	return jitter
}
