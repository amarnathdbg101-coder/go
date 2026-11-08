package timingwheel

import (
	"sync"
	"time"
)

// Timer represents an individual scheduled delayed task on the timing wheel.
type Timer struct {
	rounds    int
	slot      int
	task      func()
	cancelled bool
	prev      *Timer
	next      *Timer
	bucket    *slotBucket
}

// Stop cancels the timer if it has not yet executed.
// Returns true if the timer was successfully cancelled, false if already executed or stopped.
func (t *Timer) Stop() bool {
	if t.bucket == nil {
		return false
	}
	return t.bucket.tw.cancelTimer(t)
}

type slotBucket struct {
	tw   *TimingWheel
	head *Timer
	tail *Timer
}

func (b *slotBucket) push(t *Timer) {
	t.bucket = b
	t.prev = b.tail
	t.next = nil
	if b.tail != nil {
		b.tail.next = t
	} else {
		b.head = t
	}
	b.tail = t
}

func (b *slotBucket) remove(t *Timer) {
	if t.prev != nil {
		t.prev.next = t.next
	} else {
		b.head = t.next
	}

	if t.next != nil {
		t.next.prev = t.prev
	} else {
		b.tail = t.prev
	}

	t.prev = nil
	t.next = nil
	t.bucket = nil
}

// TimingWheel implements a circular hashed timing wheel for high-volume O(1) timer management.
type TimingWheel struct {
	mu          sync.Mutex
	tick        time.Duration
	wheelSize   int
	currentSlot int
	slots       []*slotBucket
	stopChan    chan struct{}
	stopped     bool
}

// New creates a new TimingWheel with the given tick resolution and slot count.
func New(tick time.Duration, wheelSize int) *TimingWheel {
	if tick <= 0 || wheelSize <= 0 {
		panic("timingwheel: tick and wheelSize must be greater than 0")
	}

	tw := &TimingWheel{
		tick:      tick,
		wheelSize: wheelSize,
		slots:     make([]*slotBucket, wheelSize),
		stopChan:  make(chan struct{}),
	}

	for i := 0; i < wheelSize; i++ {
		tw.slots[i] = &slotBucket{tw: tw}
	}

	return tw
}

// Start launches the background ticker goroutine.
func (tw *TimingWheel) Start() {
	go tw.run()
}

// Stop stops the timing wheel and terminates the background ticker.
func (tw *TimingWheel) Stop() {
	tw.mu.Lock()
	if tw.stopped {
		tw.mu.Unlock()
		return
	}
	tw.stopped = true
	close(tw.stopChan)
	tw.mu.Unlock()
}

// AfterFunc schedules a function to execute after duration d. Returns a Timer handle to cancel.
func (tw *TimingWheel) AfterFunc(d time.Duration, fn func()) (*Timer, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.stopped {
		return nil, ErrWheelStopped
	}

	ticks := int(d / tw.tick)
	if ticks <= 0 {
		ticks = 1
	}

	rounds := ticks / tw.wheelSize
	slot := (tw.currentSlot + (ticks % tw.wheelSize)) % tw.wheelSize

	timer := &Timer{
		rounds: rounds,
		slot:   slot,
		task:   fn,
	}

	tw.slots[slot].push(timer)
	return timer, nil
}

func (tw *TimingWheel) cancelTimer(t *Timer) bool {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if t.cancelled || t.bucket == nil {
		return false
	}

	t.cancelled = true
	t.bucket.remove(t)
	return true
}

func (tw *TimingWheel) run() {
	ticker := time.NewTicker(tw.tick)
	defer ticker.Stop()

	for {
		select {
		case <-tw.stopChan:
			return
		case <-ticker.C:
			tw.onTick()
		}
	}
}

func (tw *TimingWheel) onTick() {
	tw.mu.Lock()
	tw.currentSlot = (tw.currentSlot + 1) % tw.wheelSize
	bucket := tw.slots[tw.currentSlot]

	var toExecute []func()

	curr := bucket.head
	for curr != nil {
		next := curr.next

		if curr.cancelled {
			bucket.remove(curr)
		} else if curr.rounds > 0 {
			curr.rounds--
		} else {
			// Ready to execute
			bucket.remove(curr)
			toExecute = append(toExecute, curr.task)
		}

		curr = next
	}
	tw.mu.Unlock()

	// Execute callbacks outside lock
	for _, fn := range toExecute {
		go fn()
	}
}
