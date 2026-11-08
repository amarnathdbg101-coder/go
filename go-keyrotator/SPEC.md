# SPEC: go-keyrotator

## 1. Scope & Purpose
`go-keyrotator` provides thread-safe round-robin rotation, health telemetry, and automatic quarantine failover for API keys (Gemini, OpenAI, Anthropic), payment gateway secrets, SMS providers, and outbound proxies. It ensures applications remain resilient against HTTP 429 rate limits and transient outages.

## 2. Public API Contract
- `type KeyHealth struct`: Telemetry metadata (call counts, error streaks, rate-limit expiration).
- `type Rotator struct`: Core thread-safe pool manager.
- `func New(keys []string, opts ...Option) *Rotator`: Constructor with functional options.
- `func FromCSV(csv string, opts ...Option) *Rotator`: Helper parsing comma-separated key strings.
- `func (r *Rotator) GetNextKey() (string, error)`: Fetches next healthy key or earliest recovery candidate.
- `func (r *Rotator) MarkRateLimited(key string, cooldown time.Duration)`: Enforces cooldown on rate limit.
- `func (r *Rotator) MarkSuccess(key string)`: Clears error streaks on successful responses.
- `func (r *Rotator) MarkError(key string)`: Increments error count and auto-trips if threshold is met.

## 3. Concurrency & Memory Guarantees
- Zero background goroutine leaks: Evaluation of cooldown expiration happens lazily on `GetNextKey()`.
- Thread safety: Guarded by `sync.RWMutex`.
- Memory bounds: Fixed allocation proportional to configured key count.

## 4. How to Extend
New selection algorithms can implement the `Strategy` interface (`Select(keys []*KeyHealth) (*KeyHealth, error)`).
