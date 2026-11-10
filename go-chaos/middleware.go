package chaos

import (
	"fmt"
	"net/http"
	"time"
)

// Middleware wraps an http.Handler to inject faults dynamically.
func Middleware(engine *Engine) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rule := engine.Evaluate(r)
			if rule == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Add diagnostic header indicating a fault was injected
			w.Header().Set("X-Chaos-Injected", string(rule.Type))

			switch rule.Type {
			case FaultLatency:
				if rule.Delay > 0 {
					time.Sleep(rule.Delay)
				}
				next.ServeHTTP(w, r)

			case FaultError:
				status := rule.StatusCode
				if status == 0 {
					status = http.StatusInternalServerError
				}
				body := rule.StatusBody
				if body == "" {
					body = fmt.Sprintf(`{"error":"chaos_injected","status":%d,"rule":"%s"}`, status, rule.ID)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = w.Write([]byte(body))

			case FaultPanic:
				msg := rule.PanicMsg
				if msg == "" {
					msg = "chaos: synthetic emergency panic injected"
				}
				panic(msg)

			case FaultCorrupt:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("%%%CORRUPTED_CHAOS_PAYLOAD_GARBAGE_BYTES%%%"))

			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}
