package netkit

import "errors"

var (
	// ErrPortUnavailable is returned when a free TCP port cannot be allocated.
	ErrPortUnavailable = errors.New("netkit: could not allocate a free port")

	// ErrInvalidCIDR is returned when a CIDR block string cannot be parsed.
	ErrInvalidCIDR = errors.New("netkit: invalid cidr address")

	// ErrInvalidIP is returned when an IP address string cannot be parsed.
	ErrInvalidIP = errors.New("netkit: invalid ip address")

	// ErrTimeout is returned when waiting for a port exceeds the context timeout.
	ErrTimeout = errors.New("netkit: timeout waiting for port")
)
