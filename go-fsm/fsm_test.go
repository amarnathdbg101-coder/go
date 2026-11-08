package gofsm

import (
	"errors"
	"testing"
)

type OrderState string
type OrderEvent string

const (
	StatePending   OrderState = "PENDING"
	StatePaid      OrderState = "PAID"
	StateShipped   OrderState = "SHIPPED"
	StateDelivered OrderState = "DELIVERED"
	StateCancelled OrderState = "CANCELLED"

	EventPay     OrderEvent = "PAY"
	EventShip    OrderEvent = "SHIP"
	EventDeliver OrderEvent = "DELIVER"
	EventCancel  OrderEvent = "CANCEL"
)

func TestFSM_OrderLifecycle(t *testing.T) {
	fsm := New[OrderState, OrderEvent](StatePending)

	var hookLog []string

	// Register transitions
	fsm.AddTransition(StatePending, EventPay, StatePaid, nil)
	fsm.AddTransition(StatePaid, EventShip, StateShipped, nil)
	fsm.AddTransition(StateShipped, EventDeliver, StateDelivered, nil)
	fsm.AddTransition(StatePending, EventCancel, StateCancelled, nil)

	// Hooks
	fsm.OnExit(StatePending, func(to OrderState, event OrderEvent) {
		hookLog = append(hookLog, "exit_pending")
	})
	fsm.OnEnter(StatePaid, func(from OrderState, event OrderEvent) {
		hookLog = append(hookLog, "enter_paid")
	})
	fsm.OnTransition(func(from OrderState, event OrderEvent, to OrderState) {
		hookLog = append(hookLog, string(from)+"->"+string(to))
	})

	// Initial checks
	if fsm.Current() != StatePending {
		t.Fatalf("expected initial state PENDING, got %s", fsm.Current())
	}
	if !fsm.Can(EventPay) {
		t.Errorf("expected Can(EventPay) = true")
	}
	if fsm.Can(EventShip) {
		t.Errorf("expected Can(EventShip) = false from PENDING")
	}

	// 1. Pay
	err := fsm.Trigger(EventPay)
	if err != nil {
		t.Fatalf("Trigger(EventPay) failed: %v", err)
	}
	if fsm.Current() != StatePaid {
		t.Errorf("expected state PAID, got %s", fsm.Current())
	}

	// Check hooks executed in order
	expectedHooks := []string{"exit_pending", "enter_paid", "PENDING->PAID"}
	for i, h := range expectedHooks {
		if hookLog[i] != h {
			t.Errorf("hook %d: expected %s, got %s", i, h, hookLog[i])
		}
	}

	// 2. Ship
	_ = fsm.Trigger(EventShip)
	if fsm.Current() != StateShipped {
		t.Errorf("expected state SHIPPED, got %s", fsm.Current())
	}

	// 3. Deliver
	_ = fsm.Trigger(EventDeliver)
	if fsm.Current() != StateDelivered {
		t.Errorf("expected state DELIVERED, got %s", fsm.Current())
	}

	// 4. Invalid event from terminal state
	err = fsm.Trigger(EventPay)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition from DELIVERED, got %v", err)
	}
}

func TestFSM_GuardConditions(t *testing.T) {
	fsm := New[string, string]("DRAFT")

	allowPublish := false
	fsm.AddTransition("DRAFT", "PUBLISH", "PUBLISHED", func() bool {
		return allowPublish
	})

	// Guard is false -> Can should be false and Trigger should fail
	if fsm.Can("PUBLISH") {
		t.Errorf("expected Can to be false when guard fails")
	}
	err := fsm.Trigger("PUBLISH")
	if !errors.Is(err, ErrGuardFailed) {
		t.Errorf("expected ErrGuardFailed, got %v", err)
	}

	// Flip guard to true
	allowPublish = true
	if !fsm.Can("PUBLISH") {
		t.Errorf("expected Can to be true when guard passes")
	}
	err = fsm.Trigger("PUBLISH")
	if err != nil {
		t.Fatalf("expected trigger to succeed when guard is true, got %v", err)
	}
	if fsm.Current() != "PUBLISHED" {
		t.Errorf("expected state PUBLISHED, got %s", fsm.Current())
	}
}
