package pal

import (
	"errors"
)

var (
	ErrServiceNotFound      = errors.New("service not found")
	ErrServiceNotInit       = errors.New("service not initialized")
	ErrServiceInitFailed    = errors.New("service initialization failed")
	ErrServiceCastingFailed = errors.New("service casting failed")
)

// TODO: specify the exit codes

// RunError is returned by pal.Run() if it encounters an error, it wraps the original error and specifies the exit code.
type RunError struct {
	root error
}

func (e *RunError) Error() string {
	return e.root.Error()
}

func (e *RunError) Root() error {
	return e.root
}

func (e *RunError) ExitCode() int {
	return -1
}
