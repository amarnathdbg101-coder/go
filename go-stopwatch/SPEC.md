# Technical Specification: go-stopwatch

## 1. Domain & Purpose
`go-stopwatch` provides a high-resolution execution profiler and multi-lap stopwatch timer for Go. It allows developers to benchmark and pinpoint latency bottlenecks across multi-stage business workflows, database query runs, and microservice spans without requiring heavy external tracing agents.

## 2. Architecture & Design Principles
- **Sub-Nanosecond Monotonic Precision**: Uses Go runtime monotonic clock timestamps (`time.Now()`) to prevent clock skew inaccuracies from NTP updates.
- **Lap & Split Accounting**: Records individual stage duration alongside cumulative elapsed split duration.
- **Multi-Format Export**: Generates human-readable console breakdown tables (`Summary()`) with percentage distribution alongside structured telemetry maps (`ToMap()`).
- **Thread Safety**: Complete internal synchronization via `sync.Mutex`.
- **Zero Heavy Dependencies**: Pure Go standard library (`time`, `sync`, `fmt`, `strings`).

## 3. API Surface
- `type Lap struct { Name string; Duration, Split time.Duration }`
- `type Stopwatch struct`
- `func Start(name string) *Stopwatch`
- `func (s *Stopwatch) Lap(name string) time.Duration`
- `func (s *Stopwatch) Stop() time.Duration`
- `func (s *Stopwatch) Elapsed() time.Duration`
- `func (s *Stopwatch) Laps() []Lap`
- `func (s *Stopwatch) Summary() string`
- `func (s *Stopwatch) ToMap() map[string]any`
- `func (s *Stopwatch) Reset()`

## 4. Verification & Testing
- Unit tests verify consecutive lap and cumulative split calculation, console summary output formatting, structured telemetry map generation, and state reset.
