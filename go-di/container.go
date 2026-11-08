package godi

import (
	"fmt"
	"io"
	"reflect"
	"sync"
)

// Lifetime specifies how service instances are retained and reused.
type Lifetime int

const (
	Singleton Lifetime = iota
	Transient
	Instance
)

// Disposable defines custom cleanup hook.
type Disposable interface {
	Close() error
}

type serviceKey struct {
	typ  reflect.Type
	name string
}

type provider struct {
	lifetime Lifetime
	factory  func(c *Container) (any, error)
	instance any
	mu       sync.Mutex
}

// Container coordinates dependency registration, lifecycle, and resolution.
type Container struct {
	providers  map[serviceKey]*provider
	singletons []any
	mu         sync.RWMutex
	resolving  map[serviceKey]bool
	closed     bool
}

// New initializes an empty Container.
func New() *Container {
	return &Container{
		providers: make(map[serviceKey]*provider),
		resolving: make(map[serviceKey]bool),
	}
}

// RegisterSingleton registers a factory that produces a single cached instance.
func RegisterSingleton[T any](c *Container, factory func(c *Container) (T, error)) {
	RegisterNamedSingleton(c, "", factory)
}

// RegisterNamedSingleton registers a named singleton factory.
func RegisterNamedSingleton[T any](c *Container, name string, factory func(c *Container) (T, error)) {
	var zero T
	key := serviceKey{typ: reflect.TypeOf(&zero).Elem(), name: name}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.providers[key] = &provider{
		lifetime: Singleton,
		factory: func(cont *Container) (any, error) {
			return factory(cont)
		},
	}
}

// RegisterTransient registers a factory called each time the dependency is resolved.
func RegisterTransient[T any](c *Container, factory func(c *Container) (T, error)) {
	RegisterNamedTransient(c, "", factory)
}

// RegisterNamedTransient registers a named transient factory.
func RegisterNamedTransient[T any](c *Container, name string, factory func(c *Container) (T, error)) {
	var zero T
	key := serviceKey{typ: reflect.TypeOf(&zero).Elem(), name: name}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.providers[key] = &provider{
		lifetime: Transient,
		factory: func(cont *Container) (any, error) {
			return factory(cont)
		},
	}
}

// RegisterInstance registers an existing instance as a singleton.
func RegisterInstance[T any](c *Container, inst T) {
	var zero T
	key := serviceKey{typ: reflect.TypeOf(&zero).Elem(), name: ""}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.providers[key] = &provider{
		lifetime: Instance,
		instance: inst,
	}
	c.singletons = append(c.singletons, inst)
}

// Resolve retrieves an instance of type T.
func Resolve[T any](c *Container) (T, error) {
	return ResolveNamed[T](c, "")
}

// ResolveNamed retrieves an instance of type T with a given name.
func ResolveNamed[T any](c *Container, name string) (T, error) {
	var zero T
	key := serviceKey{typ: reflect.TypeOf(&zero).Elem(), name: name}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return zero, ErrContainerClosed
	}

	p, exists := c.providers[key]
	if !exists {
		c.mu.Unlock()
		return zero, fmt.Errorf("%w: %v (name: %q)", ErrServiceNotFound, key.typ, name)
	}

	// Check circular dependency
	if c.resolving[key] {
		c.mu.Unlock()
		return zero, fmt.Errorf("%w: %v (name: %q)", ErrCircularDependency, key.typ, name)
	}

	c.resolving[key] = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.resolving, key)
		c.mu.Unlock()
	}()

	switch p.lifetime {
	case Instance:
		return p.instance.(T), nil

	case Singleton:
		p.mu.Lock()
		defer p.mu.Unlock()

		if p.instance != nil {
			return p.instance.(T), nil
		}

		inst, err := p.factory(c)
		if err != nil {
			return zero, err
		}

		p.instance = inst
		c.mu.Lock()
		c.singletons = append(c.singletons, inst)
		c.mu.Unlock()

		return inst.(T), nil

	case Transient:
		inst, err := p.factory(c)
		if err != nil {
			return zero, err
		}
		return inst.(T), nil
	}

	return zero, ErrServiceNotFound
}

// MustResolve retrieves an instance of type T or panics on error.
func MustResolve[T any](c *Container) T {
	inst, err := Resolve[T](c)
	if err != nil {
		panic(err)
	}
	return inst
}

// Close gracefully disposes all singletons in reverse registration order.
func (c *Container) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	singletons := make([]any, len(c.singletons))
	copy(singletons, c.singletons)
	c.mu.Unlock()

	var firstErr error
	// Reverse order disposal (LIFO: teardown dependencies after consumers)
	for i := len(singletons) - 1; i >= 0; i-- {
		inst := singletons[i]
		if closer, ok := inst.(io.Closer); ok {
			if err := closer.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		} else if disp, ok := inst.(Disposable); ok {
			if err := disp.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}
