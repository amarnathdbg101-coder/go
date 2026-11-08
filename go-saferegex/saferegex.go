package saferegex

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"
)

type lruNode struct {
	key   string
	re    *regexp.Regexp
	prev  *lruNode
	next  *lruNode
}

// Engine caches compiled regular expressions and executes matches under bounded timeouts.
type Engine struct {
	mu             sync.Mutex
	cache          map[string]*lruNode
	head           *lruNode
	tail           *lruNode
	capacity       int
	defaultTimeout time.Duration
}

// Config configures the safe regex engine.
type Config struct {
	Capacity       int
	DefaultTimeout time.Duration
}

// DefaultConfig provides sensible defaults.
func DefaultConfig() Config {
	return Config{
		Capacity:       256,
		DefaultTimeout: 100 * time.Millisecond,
	}
}

// New creates a new Engine with the specified config.
func New(cfg Config) *Engine {
	if cfg.Capacity <= 0 {
		cfg.Capacity = 256
	}
	if cfg.DefaultTimeout <= 0 {
		cfg.DefaultTimeout = 100 * time.Millisecond
	}

	return &Engine{
		cache:          make(map[string]*lruNode),
		capacity:       cfg.Capacity,
		defaultTimeout: cfg.DefaultTimeout,
	}
}

// Default is the package-level default safe regex engine.
var Default = New(DefaultConfig())

// getOrCompile retrieves a compiled regexp from LRU cache or compiles it.
func (e *Engine) getOrCompile(pattern string) (*regexp.Regexp, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if node, ok := e.cache[pattern]; ok {
		e.moveToFront(node)
		return node.re, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: %s (%v)", ErrInvalidPattern, pattern, err)
	}

	node := &lruNode{
		key: pattern,
		re:  re,
	}

	if len(e.cache) >= e.capacity {
		e.evictTail()
	}

	e.cache[pattern] = node
	e.addToFront(node)

	return re, nil
}

func (e *Engine) moveToFront(node *lruNode) {
	if e.head == node {
		return
	}
	e.remove(node)
	e.addToFront(node)
}

func (e *Engine) addToFront(node *lruNode) {
	node.prev = nil
	node.next = e.head
	if e.head != nil {
		e.head.prev = node
	}
	e.head = node
	if e.tail == nil {
		e.tail = node
	}
}

func (e *Engine) remove(node *lruNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		e.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		e.tail = node.prev
	}
}

func (e *Engine) evictTail() {
	if e.tail == nil {
		return
	}
	delete(e.cache, e.tail.key)
	e.remove(e.tail)
}

// MatchString matches s against pattern with timeout protection against ReDoS.
func (e *Engine) MatchString(ctx context.Context, pattern, s string, timeout time.Duration) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	re, err := e.getOrCompile(pattern)
	if err != nil {
		return false, err
	}

	if timeout <= 0 {
		timeout = e.defaultTimeout
	}

	type matchResult struct {
		matched bool
	}

	resChan := make(chan matchResult, 1)

	go func() {
		resChan <- matchResult{matched: re.MatchString(s)}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-timer.C:
		return false, ErrTimeout
	case res := <-resChan:
		if err := ctx.Err(); err != nil {
			return false, err
		}
		return res.matched, nil
	}
}

// FindString searches for the first match in s within the timeout limit.
func (e *Engine) FindString(ctx context.Context, pattern, s string, timeout time.Duration) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	re, err := e.getOrCompile(pattern)
	if err != nil {
		return "", err
	}

	if timeout <= 0 {
		timeout = e.defaultTimeout
	}

	resChan := make(chan string, 1)

	go func() {
		resChan <- re.FindString(s)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-timer.C:
		return "", ErrTimeout
	case res := <-resChan:
		if err := ctx.Err(); err != nil {
			return "", err
		}
		return res, nil
	}
}

// ReplaceAllString replaces matches of pattern in src with repl within the timeout limit.
func (e *Engine) ReplaceAllString(ctx context.Context, pattern, src, repl string, timeout time.Duration) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	re, err := e.getOrCompile(pattern)
	if err != nil {
		return "", err
	}

	if timeout <= 0 {
		timeout = e.defaultTimeout
	}

	resChan := make(chan string, 1)

	go func() {
		resChan <- re.ReplaceAllString(src, repl)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-timer.C:
		return "", ErrTimeout
	case res := <-resChan:
		if err := ctx.Err(); err != nil {
			return "", err
		}
		return res, nil
	}
}

// Package-level forwarders to Default:

// MatchString tests if pattern matches s within timeout using the default engine.
func MatchString(pattern, s string, timeout time.Duration) (bool, error) {
	return Default.MatchString(context.Background(), pattern, s, timeout)
}

// FindString returns the first match of pattern in s using the default engine.
func FindString(pattern, s string, timeout time.Duration) (string, error) {
	return Default.FindString(context.Background(), pattern, s, timeout)
}

// ReplaceAllString replaces all matches of pattern in src with repl using the default engine.
func ReplaceAllString(pattern, src, repl string, timeout time.Duration) (string, error) {
	return Default.ReplaceAllString(context.Background(), pattern, src, repl, timeout)
}
