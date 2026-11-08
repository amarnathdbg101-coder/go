# Technical Specification: go-di

## 1. Domain & Purpose
`go-di` is an ultra-lightweight, type-safe, generic Dependency Injection container and service locator for Go. It eliminates tight coupling between layers (handlers, services, repositories, database connection pools) while avoiding heavy reflection frameworks or black-box code generators.

## 2. Architecture & Design Principles
- **Generics-Powered Resolution**: Clean `Resolve[T](container)` syntax with compile-time type inference.
- **Multiple Lifetimes**:
  - `Singleton`: Lazily instantiated once on first request, thread-safe cached.
  - `Transient`: Fresh instance created per resolution.
  - `Instance`: Pre-constructed object registered directly.
- **Cycle Detection**: Automatically tracks resolution stack and returns `ErrCircularDependency` instead of stack overflow panics.
- **LIFO Lifecycle Teardown**: `Close()` walks resolved singletons in reverse registration order, automatically invoking `io.Closer` or `Disposable`.
- **Zero Heavy Dependencies**: Pure Go standard library (`reflect`, `sync`, `fmt`).

## 3. API Surface
- `type Lifetime int` (`Singleton`, `Transient`, `Instance`)
- `type Disposable interface { Close() error }`
- `func New() *Container`
- `func RegisterSingleton[T any](c *Container, factory func(c *Container) (T, error))`
- `func RegisterNamedSingleton[T any](c *Container, name string, factory func(c *Container) (T, error))`
- `func RegisterTransient[T any](c *Container, factory func(c *Container) (T, error))`
- `func RegisterNamedTransient[T any](c *Container, name string, factory func(c *Container) (T, error))`
- `func RegisterInstance[T any](c *Container, inst T)`
- `func Resolve[T any](c *Container) (T, error)`
- `func ResolveNamed[T any](c *Container, name string) (T, error)`
- `func MustResolve[T any](c *Container) T`
- `func (c *Container) Close() error`

## 4. Error Handling
- `ErrServiceNotFound`: Requested service not found in registry.
- `ErrCircularDependency`: Cycle detected in dependency resolution graph.
- `ErrContainerClosed`: Container closed.

## 5. Verification & Testing
- Unit tests verify singleton caching, transient distinct allocations, named registrations, circular dependency errors, and LIFO `io.Closer` teardown on shutdown.
