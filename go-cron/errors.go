package gocron

import "errors"

var (
	// ErrInvalidCronExpr indicates a malformed 5-field cron expression.
	ErrInvalidCronExpr = errors.New("cron: invalid 5-field cron expression (expected: minute hour dom month dow)")

	// ErrInvalidField indicates an out-of-range value in a cron field.
	ErrInvalidField = errors.New("cron: field value out of bounds")

	// ErrSchedulerRunning occurs when calling Start on an already active scheduler.
	ErrSchedulerRunning = errors.New("cron: scheduler is already running")
)
