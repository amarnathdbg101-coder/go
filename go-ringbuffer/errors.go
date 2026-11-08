package goringbuffer

import "errors"

var (
	// ErrBufferFull indicates attempting to push into a full buffer when overwrite is disabled.
	ErrBufferFull = errors.New("goringbuffer: ring buffer is full")

	// ErrBufferEmpty indicates attempting to pop from an empty ring buffer.
	ErrBufferEmpty = errors.New("goringbuffer: ring buffer is empty")

	// ErrInvalidCapacity indicates ring buffer capacity must be greater than zero.
	ErrInvalidCapacity = errors.New("goringbuffer: capacity must be greater than zero")
)
