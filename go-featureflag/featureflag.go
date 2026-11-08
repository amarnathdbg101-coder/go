package featureflag

import (
	"hash/fnv"
	"sync"
	"time"
)

// Context provides entity data (such as UserID and custom attributes) for flag evaluation.
type Context struct {
	UserID     string
	Attributes map[string]string
}

// Flag defines a feature toggle configuration.
type Flag struct {
	Key           string
	Enabled       bool
	Percentage    int               // 0 to 100
	AllowedUsers  []string          // Whitelisted users that bypass percentage
	RequiredAttrs map[string]string // Attributes required to enable flag
	StartTime     *time.Time        // Optional active start time
	EndTime       *time.Time        // Optional expiration end time
}

// Engine manages thread-safe feature flags in memory.
type Engine struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

// New creates a new feature flag Engine.
func New() *Engine {
	return &Engine{
		flags: make(map[string]Flag),
	}
}

// Register stores or updates a flag in the engine.
func (e *Engine) Register(flag Flag) error {
	if flag.Percentage < 0 || flag.Percentage > 100 {
		return ErrInvalidPercentage
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.flags[flag.Key] = flag
	return nil
}

// SetBoolean creates or sets a simple binary on/off flag.
func (e *Engine) SetBoolean(key string, enabled bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.flags[key] = Flag{
		Key:     key,
		Enabled: enabled,
	}
}

// SetPercentage configures a percentage rollout for a flag.
func (e *Engine) SetPercentage(key string, percentage int) error {
	if percentage < 0 || percentage > 100 {
		return ErrInvalidPercentage
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	f, exists := e.flags[key]
	if !exists {
		f = Flag{Key: key}
	}
	f.Enabled = true
	f.Percentage = percentage
	e.flags[key] = f
	return nil
}

// Remove deletes a flag from the engine.
func (e *Engine) Remove(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.flags, key)
}

// IsEnabled evaluates whether a feature flag is enabled for the provided context.
func (e *Engine) IsEnabled(key string, ctx Context) bool {
	e.mu.RLock()
	flag, exists := e.flags[key]
	e.mu.RUnlock()

	if !exists || !flag.Enabled {
		return false
	}

	now := time.Now()

	// Check time windows
	if flag.StartTime != nil && now.Before(*flag.StartTime) {
		return false
	}
	if flag.EndTime != nil && now.After(*flag.EndTime) {
		return false
	}

	// Check explicit user whitelist
	if len(flag.AllowedUsers) > 0 {
		isWhitelisted := false
		if ctx.UserID != "" {
			for _, u := range flag.AllowedUsers {
				if u == ctx.UserID {
					isWhitelisted = true
					break
				}
			}
		}
		if isWhitelisted {
			return true
		}
		// If not whitelisted and no percentage rollout is defined, access is denied
		if flag.Percentage <= 0 {
			return false
		}
	}

	// Check required attributes
	if len(flag.RequiredAttrs) > 0 {
		if ctx.Attributes == nil {
			return false
		}
		for k, expectedVal := range flag.RequiredAttrs {
			actualVal, found := ctx.Attributes[k]
			if !found || actualVal != expectedVal {
				return false
			}
		}
	}

	// Check percentage rollout
	if flag.Percentage > 0 {
		if flag.Percentage >= 100 {
			return true
		}
		if ctx.UserID == "" {
			return false
		}

		h := fnv.New32a()
		h.Write([]byte(flag.Key + ":" + ctx.UserID))
		bucket := int(h.Sum32() % 100)
		return bucket < flag.Percentage
	}

	return flag.Enabled
}

// Get returns the current flag configuration.
func (e *Engine) Get(key string) (Flag, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	flag, exists := e.flags[key]
	if !exists {
		return Flag{}, ErrFlagNotFound
	}
	return flag, nil
}

// List returns a snapshot of all registered flags.
func (e *Engine) List() []Flag {
	e.mu.RLock()
	defer e.mu.RUnlock()

	list := make([]Flag, 0, len(e.flags))
	for _, f := range e.flags {
		list = append(list, f)
	}
	return list
}
