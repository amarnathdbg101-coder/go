package cache

import (
	"context"
	"encoding/json"
	"time"
)

type Engine struct {
	store Store
	group *Group
}

func New(store Store) (*Engine, error) {
	if store == nil {
		return nil, ErrStoreNil
	}
	return &Engine{
		store: store,
		group: NewGroup(),
	}, nil
}

func (e *Engine) Store() Store {
	return e.store
}

func (e *Engine) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return e.store.Get(ctx, key)
}

func (e *Engine) SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return e.store.Set(ctx, key, value, ttl)
}

func (e *Engine) Delete(ctx context.Context, key string) error {
	return e.store.Delete(ctx, key)
}

func (e *Engine) DeletePrefix(ctx context.Context, prefix string) error {
	return e.store.DeletePrefix(ctx, prefix)
}

func (e *Engine) Clear(ctx context.Context) error {
	return e.store.Clear(ctx)
}

func (e *Engine) Close() error {
	return e.store.Close()
}

func Get[T any](ctx context.Context, e *Engine, key string) (T, error) {
	var zero T
	bytes, err := e.store.Get(ctx, key)
	if err != nil {
		return zero, err
	}

	var val T
	if err := json.Unmarshal(bytes, &val); err != nil {
		return zero, err
	}
	return val, nil
}

func Set[T any](ctx context.Context, e *Engine, key string, value T, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return e.store.Set(ctx, key, bytes, ttl)
}

func GetOrSet[T any](
	ctx context.Context,
	e *Engine,
	key string,
	ttl time.Duration,
	fetcher func(ctx context.Context) (T, error),
) (T, error) {
	var zero T

	bytes, err := e.store.Get(ctx, key)
	if err == nil {
		var cachedVal T
		if err := json.Unmarshal(bytes, &cachedVal); err == nil {
			return cachedVal, nil
		}
	}

	result, err, _ := e.group.Do(key, func() (any, error) {
		if b, errCheck := e.store.Get(ctx, key); errCheck == nil {
			var cachedVal T
			if errUnmarshal := json.Unmarshal(b, &cachedVal); errUnmarshal == nil {
				return cachedVal, nil
			}
		}

		freshVal, fetchErr := fetcher(ctx)
		if fetchErr != nil {
			return zero, fetchErr
		}

		if dataBytes, marshalErr := json.Marshal(freshVal); marshalErr == nil {
			_ = e.store.Set(ctx, key, dataBytes, ttl)
		}

		return freshVal, nil
	})

	if err != nil {
		return zero, err
	}

	typed, ok := result.(T)
	if !ok {
		return zero, ErrCacheMiss
	}
	return typed, nil
}

type TypedCache[T any] struct {
	engine *Engine
}

func Typed[T any](engine *Engine) *TypedCache[T] {
	return &TypedCache[T]{engine: engine}
}

func (tc *TypedCache[T]) Get(ctx context.Context, key string) (T, error) {
	return Get[T](ctx, tc.engine, key)
}

func (tc *TypedCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	return Set[T](ctx, tc.engine, key, value, ttl)
}

func (tc *TypedCache[T]) GetOrSet(ctx context.Context, key string, ttl time.Duration, fetcher func(ctx context.Context) (T, error)) (T, error) {
	return GetOrSet[T](ctx, tc.engine, key, ttl, fetcher)
}

func (tc *TypedCache[T]) Delete(ctx context.Context, key string) error {
	return tc.engine.Delete(ctx, key)
}
