package shadow

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Middleware creates a standard net/http middleware for request mirroring.
func Middleware(engine *Engine) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !engine.ShouldSample() {
				next.ServeHTTP(w, r)
				return
			}

			// Read request body without consuming it permanently
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}

			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			start := time.Now()
			next.ServeHTTP(rec, r)
			primaryLatency := time.Since(start)

			// Enqueue shadow comparison asynchronously
			engine.Enqueue(shadowTask{
				method:         r.Method,
				path:           r.URL.Path,
				headers:        r.Header.Clone(),
				body:           bodyBytes,
				primaryStatus:  rec.statusCode,
				primaryBody:    rec.body.Bytes(),
				primaryLatency: primaryLatency,
			})
		})
	}
}
