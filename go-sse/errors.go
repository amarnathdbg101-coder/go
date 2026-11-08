package sse

import "errors"

var (
	// ErrStreamingNotSupported is returned when the ResponseWriter does not implement http.Flusher.
	ErrStreamingNotSupported = errors.New("sse: http.Flusher interface is not supported by ResponseWriter")

	// ErrStreamClosed is returned when attempting to send an event on a closed stream or cancelled request context.
	ErrStreamClosed = errors.New("sse: client connection closed")
)
