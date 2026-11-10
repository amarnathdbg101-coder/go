# SPEC: go-chaos

## 1. Scope & Purpose
`go-chaos` provides in-process chaos engineering and fault injection for Go HTTP servers, microservice clients, and resilience pipelines. It enables automated resilience verification without requiring external proxies or Kubernetes service meshes.

It does NOT provide permanent rate limiting or security filtering (use `go-ratelimit` or `go-secguard`).

## 2. Public API Contract
```go
type FaultType string
type Rule struct { ... }
type Engine struct { ... }

func New(opts ...Option) *Engine
func (e *Engine) AddRule(r *Rule) string
func (e *Engine) RemoveRule(id string) error
func (e *Engine) ClearRules()
func (e *Engine) Evaluate(req *http.Request) *Rule
func (e *Engine) SetEnabled(enabled bool)

func Middleware(engine *Engine) func(http.Handler) http.Handler
func LatencyRule(pathPrefix string, delay time.Duration, probability float64) *Rule
func ErrorRule(pathPrefix string, statusCode int, probability float64) *Rule
```

## 3. Concurrency & Memory Guarantees
- Rules map access is synchronized using `sync.RWMutex`.
- Random probability rolls use cryptographic CSPRNG without shared seed locks.
- Zero heap allocations when no rule matches a request.

## 4. How to Extend
- Add dynamic remote rule syncing from Consul, Redis, or Environment variables.
- Add gRPC interceptor support.
