package godi

import "errors"

var (
	// ErrServiceNotFound indicates the requested service type or name has not been registered.
	ErrServiceNotFound = errors.New("godi: service not registered")

	// ErrCircularDependency indicates a cycle was detected during service resolution.
	ErrCircularDependency = errors.New("godi: circular dependency detected")

	// ErrContainerClosed indicates resolution was attempted on a closed container.
	ErrContainerClosed = errors.New("godi: container is closed")
)
