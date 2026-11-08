package gopipeline

import "errors"

var (
	// ErrPipelineAborted indicates pipeline execution was aborted early.
	ErrPipelineAborted = errors.New("gopipeline: pipeline execution aborted")

	// ErrStepPanicked indicates a step encountered an unhandled panic.
	ErrStepPanicked = errors.New("gopipeline: pipeline step panicked")

	// ErrNilContext indicates a nil context was passed to execution.
	ErrNilContext = errors.New("gopipeline: context cannot be nil")
)
