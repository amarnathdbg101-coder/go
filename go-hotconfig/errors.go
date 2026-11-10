package hotconfig

import "errors"

var (
	// ErrConfigNotFound is returned when the target configuration file does not exist.
	ErrConfigNotFound = errors.New("hotconfig: configuration file not found")

	// ErrCorruptConfig is returned when the configuration file contains invalid syntax.
	ErrCorruptConfig = errors.New("hotconfig: corrupt or unparseable configuration data")

	// ErrValidationFailed is returned when the new configuration fails the user validation schema.
	ErrValidationFailed = errors.New("hotconfig: configuration failed schema validation")

	// ErrWatcherClosed is returned when attempting operations on a stopped watcher.
	ErrWatcherClosed = errors.New("hotconfig: watcher is closed")
)
