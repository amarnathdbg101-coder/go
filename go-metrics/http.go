package gometrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Handler returns an HTTP handler that exposes collected metrics to Prometheus scrapers.
func Handler(r *Registry, samplers ...func()) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Run on-demand collectors (e.g. runtime memory)
		for _, s := range samplers {
			if s != nil {
				s()
			}
		}

		output := r.Gather()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(output))
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// HTTPMiddleware automatically instruments requests with total count and latency distribution.
func HTTPMiddleware(r *Registry) func(next http.Handler) http.Handler {
	reqCounter := r.Counter("http_requests_total", "Total HTTP requests handled.", nil)
	_ = reqCounter

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()

			rec := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, req)

			duration := time.Since(start).Seconds()
			statusStr := strconv.Itoa(rec.statusCode)

			labels := Labels{
				"method": req.Method,
				"status": statusStr,
			}

			// Increment labeled counter
			r.Counter("http_requests_total", "Total HTTP requests handled.", labels).Inc()

			// Observe labeled duration
			histLabels := Labels{
				"method": req.Method,
			}
			r.Histogram("http_request_duration_seconds", "HTTP request latency in seconds.", DefaultBuckets, histLabels).Observe(duration)
			_ = fmt.Sprintf("%g", duration)
		})
	}
}
