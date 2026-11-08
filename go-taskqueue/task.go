package taskqueue

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Task represents an asynchronous unit of work dispatched through the queue.
type Task struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Payload    []byte    `json:"payload"`
	RetryCount int       `json:"retry_count"`
	MaxRetries int       `json:"max_retries"`
	CreatedAt  time.Time `json:"created_at"`
	LastError  string    `json:"last_error,omitempty"`
}

// HandlerFunc is the signature of a task processor function.
// It receives a context (with timeout) and the task instance.
// Returning nil marks the task as completed. Returning an error triggers automatic retry.
type HandlerFunc func(ctx context.Context, task *Task) error

// NewTask creates an initialized Task with JSON-encoded payload.
func NewTask(taskType string, payload any) (*Task, error) {
	var payloadBytes []byte
	if payload != nil {
		switch p := payload.(type) {
		case []byte:
			payloadBytes = p
		case string:
			payloadBytes = []byte(p)
		default:
			bytes, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("taskqueue: failed to serialize payload: %w", err)
			}
			payloadBytes = bytes
		}
	}

	return &Task{
		ID:        generateTaskID(),
		Type:      taskType,
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}, nil
}

// UnmarshalPayload deserializes the task's JSON payload into target struct.
func (t *Task) UnmarshalPayload(dst any) error {
	if len(t.Payload) == 0 {
		return nil
	}
	return json.Unmarshal(t.Payload, dst)
}

func generateTaskID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("tsk_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b))
}
