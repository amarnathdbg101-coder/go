# Technical Specification: go-cron

## 1. Domain & Purpose
`go-cron` is an ultra-lightweight, zero-external-dependency background in-memory cron and periodic job scheduler for Go applications. It provides standard 5-field Unix cron syntax parsing (`minute hour dom month dow`), custom duration intervals (`Every`), panic isolation per job, dynamic job management (add, remove, query entries), and clean graceful shutdown semantics.

## 2. Architecture & Design Principles
- **Zero Heavy Dependencies**: Pure Go standard library (`sync`, `sync/atomic`, `time`, `strings`, `strconv`, `fmt`).
- **Panic Isolation**: Individual job panics are recovered with an optional custom `OnPanic` hook, ensuring one failing job never crashes the scheduler engine or other running routines.
- **Concurrent Dispatch**: Each scheduled run executes in its own managed goroutine tracked via `sync.WaitGroup`, preventing slow-running jobs from blocking or skewing subsequent task dispatch.
- **Clean Lifecycle**: Implements `Start()`, `Stop()`, and entry inspection routines, ensuring background ticker goroutines terminate cleanly without leaking resources.

## 3. API Surface
- `type Schedule interface { Next(t time.Time) time.Time }`
- `func Every(d time.Duration) Schedule`
- `func ParseCron(expr string) (*CronSchedule, error)`
- `func New() *Scheduler`
- `func (s *Scheduler) Add(name string, schedule Schedule, job func()) uint64`
- `func (s *Scheduler) AddCron(name, cronExpr string, job func()) (uint64, error)`
- `func (s *Scheduler) Every(name string, d time.Duration, job func()) uint64`
- `func (s *Scheduler) Remove(id uint64)`
- `func (s *Scheduler) Entries() []Entry`
- `func (s *Scheduler) SetOnPanic(fn func(name string, recovered any))`
- `func (s *Scheduler) Start() error`
- `func (s *Scheduler) Stop()`

## 4. Error Handling
- `ErrInvalidCronExpr`: Malformed cron format or invalid field counts.
- `ErrInvalidField`: Out-of-bounds field value or invalid step/range syntax.
- `ErrSchedulerRunning`: Returned when `Start()` is invoked on an already running scheduler.

## 5. Verification & Testing
- Unit tests covering cron expression parsing (`*`, `*/n`, ranges `m-n`, lists `a,b,c`).
- Interval execution timing and recurrence verification.
- Panic isolation test validating scheduler continuity when a job panics.
