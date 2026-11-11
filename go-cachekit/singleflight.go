package cachekit

import "sync"

// call represents an active or completed in-flight request.
type call struct {
	wg  sync.WaitGroup
	val any
	err error
}

// Group provides request coalescing (Singleflight).
// Yeh mechanism "Cache Stampede" / "Dogpile Effect" ko 100% eliminate karta hai:
// Jab ek hi key ke liye 1,000 concurrent goroutines cache miss dekhti hain,
// toh sirf pehli goroutine database fetch run karti hai, baaki 999 goroutines wait karke
// wahi result share karti hain.
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// NewGroup creates a new singleflight group.
func NewGroup() *Group {
	return &Group{
		m: make(map[string]*call),
	}
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a time.
// If a duplicate comes in, the duplicate caller waits for the original to complete
// and receives the same results.
func (g *Group) Do(key string, fn func() (any, error)) (v any, err error, shared bool) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		// Wait for the single in-flight call to finish
		c.wg.Wait()
		return c.val, c.err, true
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	defer func() {
		g.mu.Lock()
		delete(g.m, key)
		g.mu.Unlock()
		c.wg.Done()
	}()

	// Execute fetcher function (panic-safe: defer ensures c.wg.Done() runs even if fn panics)
	c.val, c.err = fn()
	return c.val, c.err, false
}
