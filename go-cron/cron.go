package gocron

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Entry represents an active scheduled job.
type Entry struct {
	ID       uint64
	Name     string
	Schedule Schedule
	NextTime time.Time
	PrevTime time.Time
	Job      func()
}

// Scheduler coordinates concurrent scheduled cron jobs.
type Scheduler struct {
	mu       sync.Mutex
	entries  []*Entry
	nextID   uint64
	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup
	onPanic  func(name string, recovered any)
}

// New initializes a Scheduler.
func New() *Scheduler {
	return &Scheduler{
		entries:  make([]*Entry, 0),
		stopChan: make(chan struct{}),
	}
}

// SetOnPanic registers a global panic recovery callback.
func (s *Scheduler) SetOnPanic(fn func(name string, recovered any)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onPanic = fn
}

// Add schedules a job using an arbitrary Schedule implementation.
func (s *Scheduler) Add(name string, schedule Schedule, job func()) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := atomic.AddUint64(&s.nextID, 1)
	now := time.Now()

	entry := &Entry{
		ID:       id,
		Name:     name,
		Schedule: schedule,
		NextTime: schedule.Next(now),
		Job:      job,
	}

	s.entries = append(s.entries, entry)
	return id
}

// AddCron schedules a job using a standard 5-field cron expression (e.g. "*/5 * * * *").
func (s *Scheduler) AddCron(name, cronExpr string, job func()) (uint64, error) {
	schedule, err := ParseCron(cronExpr)
	if err != nil {
		return 0, err
	}
	return s.Add(name, schedule, job), nil
}

// Every schedules a job running at fixed duration intervals.
func (s *Scheduler) Every(name string, d time.Duration, job func()) uint64 {
	return s.Add(name, Every(d), job)
}

// Start launches the background scheduling loop.
func (s *Scheduler) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return ErrSchedulerRunning
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run()
	return nil
}

func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case now := <-ticker.C:
			s.triggerDueJobs(now)
		}
	}
}

func (s *Scheduler) triggerDueJobs(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Sort entries by NextTime ascending
	sort.Slice(s.entries, func(i, j int) bool {
		return s.entries[i].NextTime.Before(s.entries[j].NextTime)
	})

	for _, entry := range s.entries {
		if entry.NextTime.IsZero() || now.Before(entry.NextTime) {
			break // No further due jobs
		}

		// Fire job asynchronously to avoid blocking the scheduler tick
		go s.safeExecute(entry)

		entry.PrevTime = entry.NextTime
		entry.NextTime = entry.Schedule.Next(now)
	}
}

func (s *Scheduler) safeExecute(entry *Entry) {
	defer func() {
		if r := recover(); r != nil {
			s.mu.Lock()
			cb := s.onPanic
			s.mu.Unlock()

			if cb != nil {
				cb(entry.Name, r)
			} else {
				_ = fmt.Sprintf("cron job %s panicked: %v", entry.Name, r)
			}
		}
	}()

	entry.Job()
}

// Stop stops the scheduler and waits for active background loops to terminate.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopChan)
	s.mu.Unlock()

	s.wg.Wait()
}
