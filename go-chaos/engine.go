package chaos

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
)

// Engine manages active chaos rules and evaluates requests.
type Engine struct {
	cfg   Config
	mu    sync.RWMutex
	rules map[string]*Rule
}

// New creates a new Chaos Engine with the specified options.
func New(opts ...Option) *Engine {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Engine{
		cfg:   cfg,
		rules: make(map[string]*Rule),
	}
}

// SetEnabled toggles the master chaos killswitch.
func (e *Engine) SetEnabled(enabled bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cfg.Enabled = enabled
}

// IsEnabled returns true if the master switch is active.
func (e *Engine) IsEnabled() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.cfg.Enabled
}

// AddRule registers a new fault injection rule.
func (e *Engine) AddRule(r *Rule) string {
	e.mu.Lock()
	defer e.mu.Unlock()

	if r.ID == "" {
		var b [8]byte
		_, _ = rand.Read(b[:])
		r.ID = fmt.Sprintf("rule_%s", hex.EncodeToString(b[:]))
	}

	e.rules[r.ID] = r
	return r.ID
}

// RemoveRule deletes a rule by ID.
func (e *Engine) RemoveRule(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.rules[id]; !exists {
		return ErrRuleNotFound
	}
	delete(e.rules, id)
	return nil
}

// ClearRules removes all active rules.
func (e *Engine) ClearRules() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = make(map[string]*Rule)
}

// Evaluate checks if the incoming request matches any active chaos rule.
func (e *Engine) Evaluate(req *http.Request) *Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.cfg.Enabled {
		return nil
	}

	// Bypass header verification
	if e.cfg.BypassHeader != "" {
		val := req.Header.Get(e.cfg.BypassHeader)
		if val != "" {
			if e.cfg.BypassToken == "" || val == e.cfg.BypassToken {
				return nil
			}
		}
	}

	for _, rule := range e.rules {
		if rule.Matches(req) {
			return rule
		}
	}

	return nil
}
