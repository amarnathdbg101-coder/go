# SPEC: go-env/hotconfig

## 1. Scope & Purpose
`go-env/hotconfig` manages dynamic, zero-downtime application configuration reloading from local files or mounted Kubernetes ConfigMaps. It guarantees sub-nanosecond reads, atomic pointer swapping, and corrupt file rejection.

It does NOT manage remote distributed key-value stores (e.g. Consul/Etcd) directly.

## 2. Public API Contract
```go
type HotConfig[T any] struct { ... }
type Option[T any] func(*Options[T])

func Watch[T any](path string, options ...Option[T]) (*HotConfig[T], error)
func (h *HotConfig[T]) Get() *T
func (h *HotConfig[T]) Reload() error
func (h *HotConfig[T]) Close() error
```

## 3. Concurrency & Memory Guarantees
- Read access (`Get`) is entirely lock-free using `atomic.Pointer[T]`.
- File reads and swaps are synchronized to prevent concurrent reload races.
- Old configurations are naturally reclaimed by the Go garbage collector when references drop.

## 4. How to Extend
- Implement remote S3 / GCP Secret Manager / Vault polling adapters.
- Add structured diff logging on config swap.
