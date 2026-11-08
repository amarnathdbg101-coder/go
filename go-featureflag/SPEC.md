# Technical Specification: go-featureflag

## 1. Domain & Purpose
`go-featureflag` is a lightweight, thread-safe dynamic feature flag and canary rollout engine for Go. It enables zero-downtime feature management, percentage-based gradual rollouts, whitelist bypass, attribute segmentation, and time-bounded promotions without external SaaS lock-in or service restarts.

## 2. Architecture & Design Principles
- **Multi-Strategy Evaluation**:
  - **Boolean Toggles**: Instant binary On/Off switches.
  - **Percentage Canary**: Deterministic hashing using FNV-1a (`hash(key + ":" + UserID) % 100`) ensuring stable user bucket assignments without state persistence.
  - **Whitelist Access**: Dedicated `AllowedUsers` slice that bypasses rollout constraints.
  - **Attribute Segmentation**: Matches context attributes (e.g. `country: IN`, `tier: gold`).
  - **Time Boundaries**: Configurable `StartTime` and `EndTime` for automated promotion launches.
- **Concurrent Thread Safety**: Backed by `sync.RWMutex` allowing thousands of simultaneous evaluation checks with zero race conditions.
- **Zero Heavy Dependencies**: Pure Go standard library (`sync`, `hash/fnv`, `time`).

## 3. API Surface
- `type Context struct { UserID string; Attributes map[string]string }`
- `type Flag struct`
- `type Engine struct`
- `func New() *Engine`
- `func (e *Engine) Register(flag Flag) error`
- `func (e *Engine) SetBoolean(key string, enabled bool)`
- `func (e *Engine) SetPercentage(key string, percentage int) error`
- `func (e *Engine) IsEnabled(key string, ctx Context) bool`
- `func (e *Engine) Remove(key string)`
- `func (e *Engine) Get(key string) (Flag, error)`
- `func (e *Engine) List() []Flag`

## 4. Error Handling
- `ErrFlagNotFound`: Returned when querying a non-existent flag key.
- `ErrInvalidPercentage`: Returned when rollout percentage is outside 0-100.

## 5. Verification & Testing
- Unit tests verify binary toggling, deterministic percentage distributions (~30% on large samples), whitelist precedence, attribute filtering, expiration timestamps, and concurrency under 30 active goroutines.
