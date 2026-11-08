package multierror

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

var (
	errDatabase = errors.New("database connection lost")
	errNetwork  = errors.New("network timeout")
	errAuth     = errors.New("unauthorized token")
)

type customValidationError struct {
	field string
}

func (c *customValidationError) Error() string {
	return fmt.Sprintf("invalid field: %s", c.field)
}

func TestMultiError_Basic(t *testing.T) {
	merr := New()
	merr.Append(errDatabase)
	merr.Append(nil) // should be ignored
	merr.Append(errNetwork, errAuth)

	if merr.Len() != 3 {
		t.Fatalf("expected len 3, got %d", merr.Len())
	}

	errStr := merr.Error()
	if !strings.Contains(errStr, "3 errors occurred") {
		t.Fatalf("unexpected formatted error string: %s", errStr)
	}
	if !strings.Contains(errStr, "database connection lost") {
		t.Fatalf("missing database error in format")
	}

	// Go 1.20+ errors.Is
	if !errors.Is(merr, errDatabase) {
		t.Fatalf("expected errors.Is to match errDatabase")
	}
	if !errors.Is(merr, errNetwork) {
		t.Fatalf("expected errors.Is to match errNetwork")
	}
	if errors.Is(merr, errors.New("unrelated error")) {
		t.Fatalf("expected errors.Is to fail on unrelated error")
	}
}

func TestMultiError_ErrorsAs(t *testing.T) {
	merr := New()
	merr.Append(errDatabase)
	merr.Append(&customValidationError{field: "email"})

	var target *customValidationError
	if !errors.As(merr, &target) {
		t.Fatalf("expected errors.As to match customValidationError")
	}
	if target.field != "email" {
		t.Fatalf("expected target field email, got %s", target.field)
	}
}

func TestMultiError_ErrorOrNil(t *testing.T) {
	merr := New()
	if merr.ErrorOrNil() != nil {
		t.Fatalf("expected nil when empty")
	}

	merr.Append(errDatabase)
	if merr.ErrorOrNil() == nil {
		t.Fatalf("expected non-nil when errors exist")
	}
}

func TestMultiError_Flattening(t *testing.T) {
	merr1 := New()
	merr1.Append(errDatabase)

	merr2 := New()
	merr2.Append(errNetwork)
	merr2.Append(merr1) // nested append

	if merr2.Len() != 2 {
		t.Fatalf("expected flattened length 2, got %d", merr2.Len())
	}
}

func TestMultiError_FilterAndPrefix(t *testing.T) {
	merr := New()
	merr.Append(errDatabase, errNetwork, &customValidationError{field: "age"})

	// Filter only standard errors (exclude customValidationError)
	filtered := merr.Filter(func(err error) bool {
		var target *customValidationError
		return !errors.As(err, &target)
	})

	if filtered.Len() != 2 {
		t.Fatalf("expected 2 filtered errors, got %d", filtered.Len())
	}

	// Prefix
	prefixed := filtered.Prefix("Worker-1")
	if !strings.Contains(prefixed.Error(), "Worker-1: database connection lost") {
		t.Fatalf("unexpected prefixed error: %s", prefixed.Error())
	}
}

func TestMultiError_Concurrent(t *testing.T) {
	merr := New()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			merr.Append(fmt.Errorf("worker %d failed", id))
			_ = merr.Len()
			_ = merr.Errors()
		}(i)
	}
	wg.Wait()

	if merr.Len() != 50 {
		t.Fatalf("expected 50 accumulated errors, got %d", merr.Len())
	}
}
