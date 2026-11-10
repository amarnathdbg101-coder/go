package delayqueue

import "time"

// Driver defines the storage contract for persisting and scheduling delayed jobs.
type Driver interface {
	// Push stores a newly scheduled job.
	Push(job *Job) error

	// PopDue retrieves jobs that are due for execution (ExecuteAt <= now) and marks them as Processing.
	PopDue(now time.Time, limit int) ([]*Job, error)

	// Ack marks a job as successfully completed.
	Ack(id string) error

	// Nack records a failed execution, reschedules with retryAt, or moves to DeadLetter if retries exhausted.
	Nack(id string, retryAt time.Time, err error) error

	// Cancel cancels and removes a pending job by ID.
	Cancel(id string) error

	// Get retrieves a job by ID.
	Get(id string) (*Job, error)

	// Close releases any driver resources.
	Close() error
}
