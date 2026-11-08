# Technical Specification: go-sysinfo

## 1. Domain & Purpose
`go-sysinfo` is a zero-dependency, real-time diagnostics and telemetry collection library for Go applications. It gathers instantaneous system metrics, process metadata, runtime goroutine tallies, and memory allocator statistics for Kubernetes health endpoints, Prometheus collectors, and debug dashboards.

## 2. Architecture & Design Principles
- **Instantaneous Point-in-Time Snapshotting**: Non-blocking telemetry collection via `runtime.ReadMemStats()`, `runtime.NumGoroutine()`, and `os.Getpid()`.
- **Memory Allocation Telemetry**: Captures Alloc, TotalAlloc, Sys, HeapInuse, HeapIdle, HeapReleased, GC counts, and last GC pause duration.
- **Human-Readable Formatting**: Built-in IEC byte formatting (`FormatBytes`) producing clean `MiB` and `GiB` representations.
- **REST & JSON Ready**: Provides `ToMap()` and JSON tags on all fields for direct HTTP responses.
- **Zero Heavy Dependencies**: Pure standard library (`runtime`, `os`, `time`, `fmt`).

## 3. API Surface
- `type MemorySnapshot struct`
- `type Snapshot struct`
- `func Collect() Snapshot`
- `func FormatBytes(b uint64) string`
- `func (s Snapshot) HumanMemory() map[string]string`
- `func (s Snapshot) ToMap() map[string]any`

## 4. Error Handling
- Safe default fallbacks (e.g. unknown hostname) with zero panics.

## 5. Verification & Testing
- Unit tests verify telemetry fields, positive PID and core counts, memory metric ranges, byte formatting thresholds, and map serialization.
