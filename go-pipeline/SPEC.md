# Technical Specification: go-pipeline

## 1. Domain & Purpose
`go-pipeline` is a generic, composable, onion-model middleware and Chain of Responsibility execution engine for Go. It streamlines multi-stage business workflows such as e-commerce checkout chains, payment validations, data enrichment pipelines, and request interceptors by decoupling distinct business rules into isolated, testable pipeline stages.

## 2. Architecture & Design Principles
- **Generic Payload `T`**: Operates on any custom context struct without type assertions or reflection overhead.
- **Onion-Model Middleware**: Allows pre-execution setup, downstream invocation via `next(ctx)`, and post-execution cleanup or metric recording.
- **Immediate Short-Circuiting**: Any returned error halts subsequent stages instantly, preventing erroneous downstream executions.
- **Context-Aware**: Honors `context.Context` cancellation at every stage boundary.
- **Panic Isolation**: Recovers from unhandled step panics with standard error wrapping or an optional user-defined `PanicHandler`.
- **Zero External Dependencies**: Pure Go standard library.

## 3. API Surface
- `type NextFunc func(ctx context.Context) error`
- `type Middleware[T any] func(ctx context.Context, payload *T, next NextFunc) error`
- `type Action[T any] func(ctx context.Context, payload *T) error`
- `type PanicHandler func(r any) error`
- `type Pipeline[T any] struct`
- `func New[T any]() *Pipeline[T]`
- `func (p *Pipeline[T]) Use(mw ...Middleware[T]) *Pipeline[T]`
- `func (p *Pipeline[T]) AddStep(action Action[T]) *Pipeline[T]`
- `func (p *Pipeline[T]) SetPanicHandler(fn PanicHandler) *Pipeline[T]`
- `func (p *Pipeline[T]) Execute(ctx context.Context, payload *T) error`

## 4. Error Handling
- `ErrPipelineAborted`: Pipeline aborted early.
- `ErrStepPanicked`: Step threw an unhandled panic (wrapped).
- `ErrNilContext`: Nil context provided to `Execute`.

## 5. Verification & Testing
- Unit tests verify sequential execution, onion-model wrap-around timing/tracing, error-based short-circuiting, context cancellation, and panic recovery.
