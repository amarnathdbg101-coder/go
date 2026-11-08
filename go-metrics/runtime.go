package gometrics

import (
	"runtime"
)

// RegisterRuntimeMetrics initializes standard Go runtime gauges in the registry.
func RegisterRuntimeMetrics(r *Registry) func() {
	goroutinesGauge := r.Gauge("go_goroutines", "Number of currently executing goroutines.", nil)
	allocBytesGauge := r.Gauge("go_memstats_alloc_bytes", "Number of bytes allocated and still in use.", nil)
	sysBytesGauge := r.Gauge("go_memstats_sys_bytes", "Number of bytes obtained from system.", nil)
	gcCountGauge := r.Gauge("go_memstats_gc_completed_total", "Total number of completed GC cycles.", nil)

	// Return a sampler closure that updates values on demand
	return func() {
		goroutinesGauge.Set(float64(runtime.NumGoroutine()))

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		allocBytesGauge.Set(float64(m.Alloc))
		sysBytesGauge.Set(float64(m.Sys))
		gcCountGauge.Set(float64(m.NumGC))
	}
}
