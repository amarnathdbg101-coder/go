package shadow

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net/http"
	"sync"
	"time"
)

type shadowTask struct {
	method         string
	path           string
	headers        http.Header
	body           []byte
	primaryStatus  int
	primaryBody    []byte
	primaryLatency time.Duration
}

// Engine manages background traffic mirroring, candidate invocation, and diff analysis.
type Engine struct {
	cfg       Config
	client    *http.Client
	workQueue chan shadowTask
	wg        sync.WaitGroup
	mu        sync.Mutex
	closed    bool
}

// New creates a new traffic shadowing engine.
func New(opts ...Option) *Engine {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	e := &Engine{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		workQueue: make(chan shadowTask, cfg.MaxWorkerQueue),
	}

	// Start 4 background comparison workers
	for i := 0; i < 4; i++ {
		e.wg.Add(1)
		go e.worker()
	}

	return e
}

// ShouldSample checks if an incoming request qualifies for traffic mirroring based on SampleRate.
func (e *Engine) ShouldSample() bool {
	if e.cfg.TargetURL == "" || e.cfg.SampleRate <= 0 {
		return false
	}
	if e.cfg.SampleRate >= 1.0 {
		return true
	}

	var b [8]byte
	_, _ = rand.Read(b[:])
	val := binary.BigEndian.Uint64(b[:])
	f := float64(val) / float64(1<<64)
	return f < e.cfg.SampleRate
}

// Enqueue submits a completed primary request for shadow comparison.
func (e *Engine) Enqueue(task shadowTask) {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return
	}
	e.mu.Unlock()

	// Non-blocking submit (drop if buffer full to protect host memory)
	select {
	case e.workQueue <- task:
	default:
	}
}

// Close gracefully stops the shadow engine.
func (e *Engine) Close() error {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return nil
	}
	e.closed = true
	close(e.workQueue)
	e.mu.Unlock()

	e.wg.Wait()
	return nil
}

func (e *Engine) worker() {
	defer e.wg.Done()

	for task := range e.workQueue {
		e.processShadow(task)
	}
}

func (e *Engine) processShadow(task shadowTask) {
	targetURL := e.cfg.TargetURL + task.path
	req, err := http.NewRequestWithContext(context.Background(), task.method, targetURL, bytes.NewReader(task.body))
	if err != nil {
		return
	}

	// Replicate headers
	for k, vv := range task.headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("X-Shadowed-Traffic", "true")

	start := time.Now()
	resp, err := e.client.Do(req)
	candidateLatency := time.Since(start)

	var candidateStatus int
	var candidateBody []byte

	if err == nil {
		candidateStatus = resp.StatusCode
		candidateBody, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	} else {
		candidateStatus = http.StatusBadGateway
	}

	statusMismatch := (task.primaryStatus != candidateStatus)
	bodyMismatch := false
	if e.cfg.CompareBody && !bytes.Equal(task.primaryBody, candidateBody) {
		bodyMismatch = true
	}

	if (statusMismatch || bodyMismatch) && e.cfg.OnDiff != nil {
		diff := &DiffReport{
			Method:           task.method,
			Path:             task.path,
			PrimaryStatus:    task.primaryStatus,
			CandidateStatus:  candidateStatus,
			StatusMismatch:   statusMismatch,
			BodyMismatch:     bodyMismatch,
			PrimaryLatency:   task.primaryLatency,
			CandidateLatency: candidateLatency,
			PrimaryBody:      string(task.primaryBody),
			CandidateBody:    string(candidateBody),
			Timestamp:        time.Now(),
		}
		e.cfg.OnDiff(diff)
	}
}
