package multierror

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Formatter formats a slice of errors into a single string.
type Formatter func([]error) string

// DefaultFormatter formats errors as a bulleted indented list.
func DefaultFormatter(errs []error) string {
	if len(errs) == 0 {
		return ""
	}
	if len(errs) == 1 {
		return errs[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d errors occurred:\n", len(errs)))
	for _, err := range errs {
		sb.WriteString(fmt.Sprintf("  * %s\n", err.Error()))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// Error represents an accumulated collection of multiple errors.
// Thread-safe and compatible with Go 1.20+ Unwrap() []error, errors.Is, and errors.As.
type Error struct {
	mu        sync.RWMutex
	errs      []error
	formatter Formatter
}

// New creates a new empty MultiError accumulator.
func New(opts ...Option) *Error {
	e := &Error{
		formatter: DefaultFormatter,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Option configures an Error accumulator.
type Option func(*Error)

// WithFormatter sets a custom formatting function.
func WithFormatter(f Formatter) Option {
	return func(e *Error) {
		if f != nil {
			e.formatter = f
		}
	}
}

// Append appends non-nil errors to base. If base is not an *Error, a new *Error is created.
func Append(base error, errs ...error) error {
	var m *Error
	if base != nil {
		if me, ok := base.(*Error); ok {
			m = me
		} else {
			m = New()
			m.Append(base)
		}
	} else {
		m = New()
	}

	m.Append(errs...)
	return m.ErrorOrNil()
}

// Append safely appends errors. Ignores nil errors. If an error is already *Error, its elements are flattened.
func (e *Error) Append(errs ...error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, err := range errs {
		if err == nil {
			continue
		}
		if me, ok := err.(*Error); ok {
			me.mu.RLock()
			e.errs = append(e.errs, me.errs...)
			me.mu.RUnlock()
		} else {
			e.errs = append(e.errs, err)
		}
	}
}

// Len returns the number of accumulated errors.
func (e *Error) Len() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.errs)
}

// Errors returns a shallow copy of the accumulated errors.
func (e *Error) Errors() []error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	cp := make([]error, len(e.errs))
	copy(cp, e.errs)
	return cp
}

// ErrorOrNil returns nil if no errors exist, or e if errors were accumulated.
func (e *Error) ErrorOrNil() error {
	if e == nil {
		return nil
	}
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.errs) == 0 {
		return nil
	}
	return e
}

// Error returns the formatted string representation.
func (e *Error) Error() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.errs) == 0 {
		return ""
	}
	return e.formatter(e.errs)
}

// Unwrap returns the underlying slice of errors, enabling Go 1.20+ error tree inspection.
func (e *Error) Unwrap() []error {
	return e.Errors()
}

// Is checks if any accumulated error matches target using standard errors.Is.
func (e *Error) Is(target error) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, err := range e.errs {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// As finds the first accumulated error matching target using standard errors.As.
func (e *Error) As(target any) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, err := range e.errs {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

// Filter creates a new *Error containing only errors matching predicate fn.
func (e *Error) Filter(fn func(error) bool) *Error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	filtered := New(WithFormatter(e.formatter))
	for _, err := range e.errs {
		if fn(err) {
			filtered.errs = append(filtered.errs, err)
		}
	}
	return filtered
}

// Prefix returns a new *Error with each error message prefixed by prefix.
func (e *Error) Prefix(prefix string) *Error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	prefixed := New(WithFormatter(e.formatter))
	for _, err := range e.errs {
		prefixed.errs = append(prefixed.errs, fmt.Errorf("%s: %w", prefix, err))
	}
	return prefixed
}
