package delayqueue

import (
	"container/heap"
	"sync"
	"time"
)

type jobHeapItem struct {
	id        string
	executeAt time.Time
	index     int
}

type jobPriorityQueue []*jobHeapItem

func (pq jobPriorityQueue) Len() int           { return len(pq) }
func (pq jobPriorityQueue) Less(i, j int) bool { return pq[i].executeAt.Before(pq[j].executeAt) }
func (pq jobPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *jobPriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*jobHeapItem)
	item.index = n
	*pq = append(*pq, item)
}
func (pq *jobPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// MemoryDriver provides an in-memory priority-queue backed storage driver.
type MemoryDriver struct {
	mu    sync.Mutex
	jobs  map[string]*Job
	pq    jobPriorityQueue
	index map[string]*jobHeapItem
}

// NewMemoryDriver initializes an in-memory delay queue driver.
func NewMemoryDriver() *MemoryDriver {
	d := &MemoryDriver{
		jobs:  make(map[string]*Job),
		pq:    make(jobPriorityQueue, 0),
		index: make(map[string]*jobHeapItem),
	}
	heap.Init(&d.pq)
	return d
}

// Push adds a new job.
func (d *MemoryDriver) Push(job *Job) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.jobs[job.ID]; exists {
		return ErrDuplicateJobID
	}

	d.jobs[job.ID] = job
	item := &jobHeapItem{
		id:        job.ID,
		executeAt: job.ExecuteAt,
	}
	heap.Push(&d.pq, item)
	d.index[job.ID] = item
	return nil
}

// PopDue retrieves jobs due for execution.
func (d *MemoryDriver) PopDue(now time.Time, limit int) ([]*Job, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var due []*Job
	for d.pq.Len() > 0 && len(due) < limit {
		top := d.pq[0]
		if top.executeAt.After(now) {
			break
		}

		heap.Pop(&d.pq)
		delete(d.index, top.id)

		if job, ok := d.jobs[top.id]; ok && job.Status == StatusPending {
			job.Status = StatusProcessing
			job.UpdatedAt = now
			due = append(due, job)
		}
	}

	return due, nil
}

// Ack marks a job as completed.
func (d *MemoryDriver) Ack(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	job, exists := d.jobs[id]
	if !exists {
		return ErrJobNotFound
	}
	job.Status = StatusCompleted
	job.UpdatedAt = time.Now()
	delete(d.jobs, id)
	return nil
}

// Nack handles failed job retries or moves to DeadLetter.
func (d *MemoryDriver) Nack(id string, retryAt time.Time, err error) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	job, exists := d.jobs[id]
	if !exists {
		return ErrJobNotFound
	}

	job.CurrentRetries++
	if err != nil {
		job.LastError = err.Error()
	}
	job.UpdatedAt = time.Now()

	if job.CurrentRetries >= job.MaxRetries {
		job.Status = StatusDeadLetter
		return nil
	}

	// Reschedule for retry
	job.Status = StatusPending
	job.ExecuteAt = retryAt
	item := &jobHeapItem{
		id:        job.ID,
		executeAt: retryAt,
	}
	heap.Push(&d.pq, item)
	d.index[job.ID] = item
	return nil
}

// Cancel removes a job.
func (d *MemoryDriver) Cancel(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.jobs[id]; !exists {
		return ErrJobNotFound
	}
	if item, ok := d.index[id]; ok {
		heap.Remove(&d.pq, item.index)
		delete(d.index, id)
	}
	delete(d.jobs, id)
	return nil
}

// Get returns a job by ID.
func (d *MemoryDriver) Get(id string) (*Job, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	job, exists := d.jobs[id]
	if !exists {
		return nil, ErrJobNotFound
	}
	return job, nil
}

// Close closes the in-memory driver.
func (d *MemoryDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.jobs = nil
	d.pq = nil
	d.index = nil
	return nil
}
