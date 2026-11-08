package atomic

import (
	"encoding/json"
	"math"
	"sync/atomic"
	"time"
)

// Bool is an atomic boolean wrapper with toggle and CAS operations.
type Bool struct {
	v uint32
}

// NewBool creates a new atomic Bool with an initial value.
func NewBool(initial bool) *Bool {
	b := &Bool{}
	if initial {
		b.v = 1
	}
	return b
}

// Load atomically reads the boolean value.
func (b *Bool) Load() bool {
	return atomic.LoadUint32(&b.v) == 1
}

// Store atomically sets the boolean value.
func (b *Bool) Store(val bool) {
	var n uint32
	if val {
		n = 1
	}
	atomic.StoreUint32(&b.v, n)
}

// Swap atomically swaps the current value with val and returns the old value.
func (b *Bool) Swap(val bool) bool {
	var n uint32
	if val {
		n = 1
	}
	return atomic.SwapUint32(&b.v, n) == 1
}

// CAS executes an atomic Compare-And-Swap operation.
func (b *Bool) CAS(oldVal, newVal bool) bool {
	var o, n uint32
	if oldVal {
		o = 1
	}
	if newVal {
		n = 1
	}
	return atomic.CompareAndSwapUint32(&b.v, o, n)
}

// Toggle atomically flips the boolean value and returns the new value.
func (b *Bool) Toggle() bool {
	for {
		curr := b.Load()
		if b.CAS(curr, !curr) {
			return !curr
		}
	}
}

// MarshalJSON implements json.Marshaler.
func (b *Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Load())
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bool) UnmarshalJSON(data []byte) error {
	var v bool
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	b.Store(v)
	return nil
}

// String provides a race-free atomic string container.
type String struct {
	v atomic.Value
}

// NewString creates a new atomic String.
func NewString(initial string) *String {
	s := &String{}
	s.v.Store(initial)
	return s
}

// Load atomically reads the string.
func (s *String) Load() string {
	val := s.v.Load()
	if val == nil {
		return ""
	}
	return val.(string)
}

// Store atomically replaces the string.
func (s *String) Store(val string) {
	s.v.Store(val)
}

// Swap atomically replaces the string and returns the previous value.
func (s *String) Swap(val string) string {
	old := s.v.Swap(val)
	if old == nil {
		return ""
	}
	return old.(string)
}

// Duration provides atomic time.Duration operations.
type Duration struct {
	v int64
}

// NewDuration creates a new atomic Duration.
func NewDuration(initial time.Duration) *Duration {
	return &Duration{v: int64(initial)}
}

// Load atomically reads the duration.
func (d *Duration) Load() time.Duration {
	return time.Duration(atomic.LoadInt64(&d.v))
}

// Store atomically sets the duration.
func (d *Duration) Store(val time.Duration) {
	atomic.StoreInt64(&d.v, int64(val))
}

// Add atomically increments the duration and returns the new value.
func (d *Duration) Add(delta time.Duration) time.Duration {
	return time.Duration(atomic.AddInt64(&d.v, int64(delta)))
}

// Time provides atomic time.Time container operations.
type Time struct {
	v atomic.Value
}

// NewTime creates a new atomic Time.
func NewTime(initial time.Time) *Time {
	t := &Time{}
	t.v.Store(initial)
	return t
}

// Load atomically reads the timestamp.
func (t *Time) Load() time.Time {
	val := t.v.Load()
	if val == nil {
		return time.Time{}
	}
	return val.(time.Time)
}

// Store atomically sets the timestamp.
func (t *Time) Store(val time.Time) {
	t.v.Store(val)
}

// Float64 provides IEEE 754 atomic floating-point operations.
type Float64 struct {
	v uint64
}

// NewFloat64 creates a new atomic Float64.
func NewFloat64(initial float64) *Float64 {
	return &Float64{v: math.Float64bits(initial)}
}

// Load atomically reads the float64 value.
func (f *Float64) Load() float64 {
	return math.Float64frombits(atomic.LoadUint64(&f.v))
}

// Store atomically sets the float64 value.
func (f *Float64) Store(val float64) {
	atomic.StoreUint64(&f.v, math.Float64bits(val))
}

// Add atomically increments the float64 and returns the new value.
func (f *Float64) Add(delta float64) float64 {
	for {
		oldBits := atomic.LoadUint64(&f.v)
		oldVal := math.Float64frombits(oldBits)
		newVal := oldVal + delta
		newBits := math.Float64bits(newVal)
		if atomic.CompareAndSwapUint64(&f.v, oldBits, newBits) {
			return newVal
		}
	}
}

// CAS executes an atomic Compare-And-Swap on float64 values.
func (f *Float64) CAS(oldVal, newVal float64) bool {
	return atomic.CompareAndSwapUint64(&f.v, math.Float64bits(oldVal), math.Float64bits(newVal))
}

// Value provides a type-safe generic atomic container for any type T.
type Value[T any] struct {
	v atomic.Value
}

// NewValue creates a generic atomic value container.
func NewValue[T any](initial T) *Value[T] {
	val := &Value[T]{}
	val.v.Store(initial)
	return val
}

// Load atomically retrieves the contained value.
func (val *Value[T]) Load() T {
	res := val.v.Load()
	if res == nil {
		var zero T
		return zero
	}
	return res.(T)
}

// Store atomically replaces the contained value.
func (val *Value[T]) Store(newVal T) {
	val.v.Store(newVal)
}

// Swap atomically replaces the value and returns the old value.
func (val *Value[T]) Swap(newVal T) T {
	old := val.v.Swap(newVal)
	if old == nil {
		var zero T
		return zero
	}
	return old.(T)
}
