package gowebhook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DeliveryResult encapsulates metadata and outcome of a webhook delivery attempt.
type DeliveryResult struct {
	StatusCode    int
	Attempts      int
	Duration      time.Duration
	ResponseBody  []byte
	Success       bool
	Error         error
}

// Config controls dispatcher execution policies.
type Config struct {
	Timeout         time.Duration
	MaxRetries      int
	BackoffBase     time.Duration
	SignatureHeader string
	EventHeader     string
	IDHeader        string
	UserAgent       string
	HTTPClient      *http.Client
	WorkerPoolSize  int
}

// DefaultConfig provides recommended production settings.
func DefaultConfig() Config {
	return Config{
		Timeout:         5 * time.Second,
		MaxRetries:      3,
		BackoffBase:     50 * time.Millisecond,
		SignatureHeader: "X-Webhook-Signature",
		EventHeader:     "X-Webhook-Event",
		IDHeader:        "X-Webhook-ID",
		UserAgent:       "Go-Webhook-Dispatcher/1.0",
		HTTPClient:      &http.Client{Timeout: 5 * time.Second},
		WorkerPoolSize:  10,
	}
}

type asyncJob struct {
	ctx        context.Context
	targetURL  string
	event      string
	webhookID  string
	payload    []byte
	secret     []byte
	onComplete func(*DeliveryResult)
}

// Dispatcher manages synchronous and asynchronous outbound webhook deliveries.
type Dispatcher struct {
	cfg     Config
	jobChan chan asyncJob
	wg      sync.WaitGroup
	closed  bool
	mu      sync.Mutex
}

// NewDispatcher initializes a dispatcher with configuration and starts async worker pool.
func NewDispatcher(cfg Config) *Dispatcher {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.BackoffBase <= 0 {
		cfg.BackoffBase = 50 * time.Millisecond
	}
	if cfg.SignatureHeader == "" {
		cfg.SignatureHeader = "X-Webhook-Signature"
	}
	if cfg.EventHeader == "" {
		cfg.EventHeader = "X-Webhook-Event"
	}
	if cfg.IDHeader == "" {
		cfg.IDHeader = "X-Webhook-ID"
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "Go-Webhook-Dispatcher/1.0"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: cfg.Timeout}
	}
	if cfg.WorkerPoolSize <= 0 {
		cfg.WorkerPoolSize = 10
	}

	d := &Dispatcher{
		cfg:     cfg,
		jobChan: make(chan asyncJob, cfg.WorkerPoolSize*4),
	}

	for i := 0; i < cfg.WorkerPoolSize; i++ {
		d.wg.Add(1)
		go d.worker()
	}

	return d
}

func (d *Dispatcher) worker() {
	defer d.wg.Done()
	for job := range d.jobChan {
		res, _ := d.Dispatch(job.ctx, job.targetURL, job.event, job.webhookID, job.payload, job.secret)
		if job.onComplete != nil {
			job.onComplete(res)
		}
	}
}

// Dispatch synchronously delivers a webhook with automatic retries and HMAC signing.
func (d *Dispatcher) Dispatch(ctx context.Context, targetURL, event, webhookID string, payload []byte, secret []byte) (*DeliveryResult, error) {
	start := time.Now()
	res := &DeliveryResult{}

	sigHeader, err := Sign(secret, payload, time.Now().Unix())
	if err != nil {
		res.Error = err
		return res, err
	}

	var lastErr error
	maxAttempts := d.cfg.MaxRetries + 1

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		res.Attempts = attempt

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
		if err != nil {
			res.Error = err
			return res, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", d.cfg.UserAgent)
		req.Header.Set(d.cfg.SignatureHeader, sigHeader)
		if event != "" {
			req.Header.Set(d.cfg.EventHeader, event)
		}
		if webhookID != "" {
			req.Header.Set(d.cfg.IDHeader, webhookID)
		}

		resp, err := d.cfg.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
		} else {
			res.StatusCode = resp.StatusCode
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
			_ = resp.Body.Close()
			res.ResponseBody = body

			// HTTP 2xx indicates success
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				res.Success = true
				res.Duration = time.Since(start)
				return res, nil
			}

			// 4xx client errors (e.g. 400 Bad Request, 401, 404) are not retried except 429 Too Many Requests
			if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
				res.Error = fmt.Errorf("webhook endpoint returned non-retryable status %d", resp.StatusCode)
				res.Duration = time.Since(start)
				return res, res.Error
			}

			lastErr = fmt.Errorf("webhook endpoint returned server error %d", resp.StatusCode)
		}

		// Wait before retry if attempts remain
		if attempt < maxAttempts {
			backoff := d.cfg.BackoffBase * (1 << (attempt - 1))
			select {
			case <-ctx.Done():
				res.Error = ctx.Err()
				res.Duration = time.Since(start)
				return res, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	res.Error = fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, lastErr)
	res.Duration = time.Since(start)
	return res, res.Error
}

// DispatchAsync queues a webhook delivery in the background pool.
func (d *Dispatcher) DispatchAsync(ctx context.Context, targetURL, event, webhookID string, payload []byte, secret []byte, onComplete func(*DeliveryResult)) {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		if onComplete != nil {
			onComplete(&DeliveryResult{Error: fmt.Errorf("dispatcher stopped")})
		}
		return
	}
	d.mu.Unlock()

	d.jobChan <- asyncJob{
		ctx:        ctx,
		targetURL:  targetURL,
		event:      event,
		webhookID:  webhookID,
		payload:    payload,
		secret:     secret,
		onComplete: onComplete,
	}
}

// Stop flushes queue and cleanly terminates background workers.
func (d *Dispatcher) Stop() {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return
	}
	d.closed = true
	d.mu.Unlock()

	close(d.jobChan)
	d.wg.Wait()
}
