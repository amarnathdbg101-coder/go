package gopipeline

import (
	"context"
	"fmt"
)

// NextFunc invokes the next step in the pipeline chain.
type NextFunc func(ctx context.Context) error

// Middleware defines an onion-style pipeline step allowing pre-processing,
// post-processing, and short-circuiting.
type Middleware[T any] func(ctx context.Context, payload *T, next NextFunc) error

// Action defines a simple sequential step that executes and moves forward.
type Action[T any] func(ctx context.Context, payload *T) error

// PanicHandler allows custom handling when a step encounters a panic.
type PanicHandler func(r any) error

// Pipeline coordinates composable execution chains over type T.
type Pipeline[T any] struct {
	middlewares []Middleware[T]
	panicHandler PanicHandler
}

// New creates an empty Pipeline for payload type T.
func New[T any]() *Pipeline[T] {
	return &Pipeline[T]{
		middlewares: make([]Middleware[T], 0),
	}
}

// Use appends one or more onion-style middleware steps.
func (p *Pipeline[T]) Use(mw ...Middleware[T]) *Pipeline[T] {
	p.middlewares = append(p.middlewares, mw...)
	return p
}

// AddStep appends a simple action step that automatically calls next.
func (p *Pipeline[T]) AddStep(action Action[T]) *Pipeline[T] {
	return p.Use(func(ctx context.Context, payload *T, next NextFunc) error {
		if err := action(ctx, payload); err != nil {
			return err
		}
		return next(ctx)
	})
}

// SetPanicHandler registers custom recovery logic.
func (p *Pipeline[T]) SetPanicHandler(fn PanicHandler) *Pipeline[T] {
	p.panicHandler = fn
	return p
}

// Execute runs the pipeline starting with the outer-most middleware.
func (p *Pipeline[T]) Execute(ctx context.Context, payload *T) (err error) {
	if ctx == nil {
		return ErrNilContext
	}

	defer func() {
		if r := recover(); r != nil {
			if p.panicHandler != nil {
				err = p.panicHandler(r)
			} else {
				err = fmt.Errorf("%w: %v", ErrStepPanicked, r)
			}
		}
	}()

	var runner func(index int, currentCtx context.Context) error
	runner = func(index int, currentCtx context.Context) error {
		// Context cancellation check
		select {
		case <-currentCtx.Done():
			return currentCtx.Err()
		default:
		}

		if index >= len(p.middlewares) {
			return nil
		}

		step := p.middlewares[index]
		return step(currentCtx, payload, func(nextCtx context.Context) error {
			return runner(index+1, nextCtx)
		})
	}

	return runner(0, ctx)
}
