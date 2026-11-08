package jsonstream

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Result wraps an item streamed over a channel alongside any decoding error.
type Result[T any] struct {
	Value T
	Err   error
}

// StreamArray reads a JSON array from r and decodes each element one-by-one into type T.
// Processing halts if fn returns an error.
func StreamArray[T any](r io.Reader, fn func(item T) error) error {
	dec := json.NewDecoder(r)

	// Read initial token; must be '['
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMalformedJSON, err)
	}

	delim, ok := tok.(json.Delim)
	if !ok || delim != '[' {
		return ErrExpectedArray
	}

	for dec.More() {
		var item T
		if err := dec.Decode(&item); err != nil {
			return fmt.Errorf("%w: %v", ErrMalformedJSON, err)
		}

		if err := fn(item); err != nil {
			return err
		}
	}

	// Consume closing ']'
	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("%w: %v", ErrMalformedJSON, err)
	}

	return nil
}

// StreamArrayChan streams elements from a JSON array over a channel.
// The channel is closed when parsing completes or context is cancelled.
func StreamArrayChan[T any](ctx context.Context, r io.Reader, bufSize int) <-chan Result[T] {
	if bufSize <= 0 {
		bufSize = 1
	}
	out := make(chan Result[T], bufSize)

	go func() {
		defer close(out)
		dec := json.NewDecoder(r)

		tok, err := dec.Token()
		if err != nil {
			out <- Result[T]{Err: fmt.Errorf("%w: %v", ErrMalformedJSON, err)}
			return
		}

		delim, ok := tok.(json.Delim)
		if !ok || delim != '[' {
			out <- Result[T]{Err: ErrExpectedArray}
			return
		}

		for dec.More() {
			var item T
			if err := dec.Decode(&item); err != nil {
				out <- Result[T]{Err: fmt.Errorf("%w: %v", ErrMalformedJSON, err)}
				return
			}

			select {
			case <-ctx.Done():
				out <- Result[T]{Err: ctx.Err()}
				return
			case out <- Result[T]{Value: item}:
			}
		}

		// Consume closing ']'
		dec.Token()
	}()

	return out
}

// StreamNDJSON decodes newline-delimited JSON (NDJSON) lines one by one.
func StreamNDJSON[T any](r io.Reader, fn func(item T) error) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var item T
		if err := json.Unmarshal(line, &item); err != nil {
			return fmt.Errorf("%w: %v", ErrMalformedJSON, err)
		}

		if err := fn(item); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// StreamNDJSONChan streams elements from an NDJSON source over a channel.
func StreamNDJSONChan[T any](ctx context.Context, r io.Reader, bufSize int) <-chan Result[T] {
	if bufSize <= 0 {
		bufSize = 1
	}
	out := make(chan Result[T], bufSize)

	go func() {
		defer close(out)
		scanner := bufio.NewScanner(r)

		for scanner.Scan() {
			line := bytes.TrimSpace(scanner.Bytes())
			if len(line) == 0 {
				continue
			}

			var item T
			if err := json.Unmarshal(line, &item); err != nil {
				out <- Result[T]{Err: fmt.Errorf("%w: %v", ErrMalformedJSON, err)}
				return
			}

			select {
			case <-ctx.Done():
				out <- Result[T]{Err: ctx.Err()}
				return
			case out <- Result[T]{Value: item}:
			}
		}

		if err := scanner.Err(); err != nil {
			out <- Result[T]{Err: err}
		}
	}()

	return out
}

// ArrayWriter writes items into an outgoing stream formatted as a valid JSON array.
type ArrayWriter[T any] struct {
	mu      sync.Mutex
	w       io.Writer
	encoder *json.Encoder
	first   bool
	closed  bool
}

// NewArrayWriter creates a new streaming JSON array writer.
func NewArrayWriter[T any](w io.Writer) *ArrayWriter[T] {
	return &ArrayWriter[T]{
		w:       w,
		encoder: json.NewEncoder(w),
		first:   true,
	}
}

// Write writes a single item into the JSON array.
func (aw *ArrayWriter[T]) Write(item T) error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	if aw.closed {
		return ErrWriterClosed
	}

	if aw.first {
		if _, err := aw.w.Write([]byte("[")); err != nil {
			return err
		}
		aw.first = false
	} else {
		if _, err := aw.w.Write([]byte(",")); err != nil {
			return err
		}
	}

	return aw.encoder.Encode(item)
}

// Close flushes the stream and writes the closing ']' character.
func (aw *ArrayWriter[T]) Close() error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	if aw.closed {
		return nil
	}
	aw.closed = true

	if aw.first {
		// Empty array
		_, err := aw.w.Write([]byte("[]"))
		return err
	}

	_, err := aw.w.Write([]byte("]"))
	return err
}

// NDJSONWriter streams JSON objects separated by newlines.
type NDJSONWriter[T any] struct {
	mu      sync.Mutex
	encoder *json.Encoder
}

// NewNDJSONWriter creates an NDJSON writer.
func NewNDJSONWriter[T any](w io.Writer) *NDJSONWriter[T] {
	return &NDJSONWriter[T]{
		encoder: json.NewEncoder(w),
	}
}

// Write serializes item as a JSON line followed by a newline.
func (nw *NDJSONWriter[T]) Write(item T) error {
	nw.mu.Lock()
	defer nw.mu.Unlock()
	return nw.encoder.Encode(item)
}
