package sysinfo

import "errors"

var (
	// ErrCollectionFailed is returned when an underlying runtime query fails.
	ErrCollectionFailed = errors.New("sysinfo: failed to collect system information")
)
