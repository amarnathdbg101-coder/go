package gohttpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Client is a resilient, production-ready HTTP client wrapper.
type Client struct {
	httpClient     *http.Client
	retryConfig    RetryConfig
	circuitBreaker *CircuitBreaker
	defaultHeaders map[string]string
}

// Option configures the Client.
type Option func(*Client)

// WithTimeout sets the overall client request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithRetry sets custom retry parameters.
func WithRetry(cfg RetryConfig) Option {
	return func(c *Client) {
		c.retryConfig = cfg
	}
}

// WithCircuitBreaker enables and configures the circuit breaker.
func WithCircuitBreaker(maxFailures int, cooldown time.Duration) Option {
	return func(c *Client) {
		c.circuitBreaker = NewCircuitBreaker(maxFailures, cooldown)
	}
}

// WithHeader sets a default header across all requests.
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.defaultHeaders[key] = value
	}
}

// New constructs a Client with production-grade HTTP connection pooling defaults.
func New(opts ...Option) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	c := &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   15 * time.Second,
		},
		retryConfig:    DefaultRetryConfig,
		circuitBreaker: nil,
		defaultHeaders: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Get performs an HTTP GET request and decodes the JSON response into target if provided.
func (c *Client) Get(ctx context.Context, urlStr string, target any) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("httpclient: build get request: %w", err)
	}
	return c.Do(req, target)
}

// Post performs an HTTP POST request with JSON encoded body and decodes JSON into target.
func (c *Client) Post(ctx context.Context, urlStr string, body any, target any) (*http.Response, error) {
	return c.sendJSON(ctx, http.MethodPost, urlStr, body, target)
}

// Put performs an HTTP PUT request with JSON encoded body and decodes JSON into target.
func (c *Client) Put(ctx context.Context, urlStr string, body any, target any) (*http.Response, error) {
	return c.sendJSON(ctx, http.MethodPut, urlStr, body, target)
}

// Delete performs an HTTP DELETE request.
func (c *Client) Delete(ctx context.Context, urlStr string, target any) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("httpclient: build delete request: %w", err)
	}
	return c.Do(req, target)
}

func (c *Client) sendJSON(ctx context.Context, method, urlStr string, body any, target any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("httpclient: marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("httpclient: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.Do(req, target)
}

// Do executes an HTTP request with circuit breaker, retries, and automatic body management.
func (c *Client) Do(req *http.Request, target any) (*http.Response, error) {
	// Set default headers
	for k, v := range c.defaultHeaders {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	// Buffer body if retries might be needed
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("httpclient: read request body: %w", err)
		}
	}

	var (
		lastResp *http.Response
		lastErr  error
	)

	maxAttempts := c.retryConfig.MaxRetries + 1
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 1. Check Circuit Breaker
		if c.circuitBreaker != nil {
			if cbErr := c.circuitBreaker.AllowRequest(); cbErr != nil {
				return nil, cbErr
			}
		}

		// Rebuild body for retry attempt
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if c.circuitBreaker != nil {
				c.circuitBreaker.RecordFailure()
			}

			if c.retryConfig.IsRetryable(0, err) && attempt < maxAttempts-1 {
				backoff := c.retryConfig.BackoffDuration(attempt)
				select {
				case <-req.Context().Done():
					return nil, req.Context().Err()
				case <-time.After(backoff):
					continue
				}
			}
			return nil, err
		}

		// Check status code for retryable conditions
		if c.retryConfig.IsRetryable(resp.StatusCode, nil) && attempt < maxAttempts-1 {
			// Drain and close body before retry to prevent connection leak
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

			if c.circuitBreaker != nil {
				c.circuitBreaker.RecordFailure()
			}

			backoff := c.retryConfig.BackoffDuration(attempt)
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(backoff):
				continue
			}
		}

		// Success or non-retryable response
		if c.circuitBreaker != nil {
			if resp.StatusCode < 500 {
				c.circuitBreaker.RecordSuccess()
			} else {
				c.circuitBreaker.RecordFailure()
			}
		}

		lastResp = resp
		break
	}

	if lastResp == nil {
		return nil, fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, lastErr)
	}

	// If HTTP status is >= 400 and target was expected
	if lastResp.StatusCode >= 400 {
		respBytes, _ := io.ReadAll(io.LimitReader(lastResp.Body, 1024))
		_ = lastResp.Body.Close()
		return lastResp, &StatusError{
			StatusCode: lastResp.StatusCode,
			Body:       string(respBytes),
		}
	}

	// Auto decode target
	if target != nil {
		defer func() {
			_, _ = io.Copy(io.Discard, lastResp.Body)
			_ = lastResp.Body.Close()
		}()

		if err := json.NewDecoder(lastResp.Body).Decode(target); err != nil {
			return lastResp, fmt.Errorf("httpclient: decode response: %w", err)
		}
	}

	return lastResp, nil
}

// AddQueryParams is a helper to append query parameters cleanly to a URL string.
func AddQueryParams(baseURL string, params map[string]string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
