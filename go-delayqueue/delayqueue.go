// Package delayqueue provides production-grade delayed job scheduling, exponential retries,
// and dead-letter queue isolation for Go.
package delayqueue

import (
	"encoding/json"
	"time"
)

// Version of the go-delayqueue package.
const Version = "1.0.0"

// EnqueueJSON marshals payload to JSON and enqueues the job.
func EnqueueJSON[T any](q *Queue, topic string, payload T, delay time.Duration) (string, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return q.Enqueue(topic, bytes, delay)
}
