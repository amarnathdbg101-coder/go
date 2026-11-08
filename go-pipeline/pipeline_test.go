package gopipeline

import (
	"context"
	"errors"
	"testing"
)

type OrderContext struct {
	ID        string
	Amount    int
	Trace     []string
	Validated bool
	Charged   bool
}

func TestPipeline_SequentialAndOnion(t *testing.T) {
	p := New[OrderContext]()

	// Middleware 1: Onion model tracing pre & post
	p.Use(func(ctx context.Context, order *OrderContext, next NextFunc) error {
		order.Trace = append(order.Trace, "mw1_start")
		err := next(ctx)
		order.Trace = append(order.Trace, "mw1_end")
		return err
	})

	// Middleware 2: AddStep action
	p.AddStep(func(ctx context.Context, order *OrderContext) error {
		order.Trace = append(order.Trace, "step_validation")
		order.Validated = true
		return nil
	})

	// Middleware 3: Final step
	p.AddStep(func(ctx context.Context, order *OrderContext) error {
		order.Trace = append(order.Trace, "step_charge")
		order.Charged = true
		return nil
	})

	order := &OrderContext{ID: "ORD_123", Amount: 100}
	err := p.Execute(context.Background(), order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !order.Validated || !order.Charged {
		t.Errorf("steps failed to execute fully: validated=%v, charged=%v", order.Validated, order.Charged)
	}

	expectedTrace := []string{"mw1_start", "step_validation", "step_charge", "mw1_end"}
	if len(order.Trace) != len(expectedTrace) {
		t.Fatalf("trace length mismatch: expected %d, got %d", len(expectedTrace), len(order.Trace))
	}
	for i, v := range expectedTrace {
		if order.Trace[i] != v {
			t.Errorf("trace[%d]: expected %s, got %s", i, v, order.Trace[i])
		}
	}
}

func TestPipeline_ShortCircuitOnError(t *testing.T) {
	p := New[OrderContext]()

	p.AddStep(func(ctx context.Context, order *OrderContext) error {
		order.Trace = append(order.Trace, "step_1")
		return errors.New("insufficient inventory")
	})

	p.AddStep(func(ctx context.Context, order *OrderContext) error {
		order.Trace = append(order.Trace, "step_2_should_not_run")
		return nil
	})

	order := &OrderContext{}
	err := p.Execute(context.Background(), order)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if len(order.Trace) != 1 || order.Trace[0] != "step_1" {
		t.Errorf("expected downstream step to be skipped on error, trace: %v", order.Trace)
	}
}

func TestPipeline_ContextCancellation(t *testing.T) {
	p := New[OrderContext]()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before execution

	p.AddStep(func(ctx context.Context, order *OrderContext) error {
		order.Trace = append(order.Trace, "should_not_execute")
		return nil
	})

	order := &OrderContext{}
	err := p.Execute(ctx, order)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestPipeline_PanicRecovery(t *testing.T) {
	p := New[OrderContext]()

	p.AddStep(func(ctx context.Context, order *OrderContext) error {
		panic("nil pointer dereference simulation")
	})

	order := &OrderContext{}
	err := p.Execute(context.Background(), order)
	if err == nil || !errors.Is(err, ErrStepPanicked) {
		t.Errorf("expected ErrStepPanicked, got %v", err)
	}

	// Custom panic handler test
	customHandled := false
	p.SetPanicHandler(func(r any) error {
		customHandled = true
		return errors.New("custom panic recovery")
	})

	err = p.Execute(context.Background(), order)
	if !customHandled || err == nil || err.Error() != "custom panic recovery" {
		t.Errorf("custom panic handler failed: %v", err)
	}
}
