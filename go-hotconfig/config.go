package hotconfig

import (
	"encoding/json"
	"time"
)

// UnmarshalerFunc parses raw bytes into a typed configuration struct pointer.
type UnmarshalerFunc[T any] func(data []byte) (*T, error)

// ValidatorFunc validates business logic and constraints of the parsed configuration.
type ValidatorFunc[T any] func(cfg *T) error

// ChangeCallback is invoked after a successful atomic configuration pointer swap.
type ChangeCallback[T any] func(oldCfg, newCfg *T)

// Options specifies configuration settings for the HotConfig watcher.
type Options[T any] struct {
	PollInterval time.Duration
	Unmarshaler  UnmarshalerFunc[T]
	Validator    ValidatorFunc[T]
	OnChange     ChangeCallback[T]
}

// Option is a functional option for configuring HotConfig.
type Option[T any] func(*Options[T])

// DefaultOptions returns production defaults for hot reloading.
func DefaultOptions[T any]() Options[T] {
	return Options[T]{
		PollInterval: 500 * time.Millisecond,
		Unmarshaler: func(data []byte) (*T, error) {
			var target T
			if err := json.Unmarshal(data, &target); err != nil {
				return nil, err
			}
			return &target, nil
		},
	}
}

// WithPollInterval sets the background polling check interval.
func WithPollInterval[T any](d time.Duration) Option[T] {
	return func(o *Options[T]) {
		if d > 0 {
			o.PollInterval = d
		}
	}
}

// WithValidator specifies a validation schema callback.
func WithValidator[T any](fn ValidatorFunc[T]) Option[T] {
	return func(o *Options[T]) {
		o.Validator = fn
	}
}

// WithUnmarshaler specifies a custom deserializer (e.g. YAML or TOML).
func WithUnmarshaler[T any](fn UnmarshalerFunc[T]) Option[T] {
	return func(o *Options[T]) {
		o.Unmarshaler = fn
	}
}

// WithOnChange registers a callback triggered on atomic configuration swap.
func WithOnChange[T any](fn ChangeCallback[T]) Option[T] {
	return func(o *Options[T]) {
		o.OnChange = fn
	}
}
