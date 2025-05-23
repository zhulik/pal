package pal

import (
	"errors"
)

// Error variables used throughout the package
var (
	// ErrServiceNotFound is returned when a requested service is not found in the container.
	// This typically happens when trying to Invoke a service that hasn't been registered.
	ErrServiceNotFound = errors.New("service not found")

	// ErrServiceInitFailed is returned when a service fails to initialize.
	// This can happen during container initialization if a service's Init method returns an error.
	ErrServiceInitFailed = errors.New("service initialization failed")

	// ErrServiceInvalid is returned when a service is invalid.
	// This can happen when a service doesn't implement a required interface or when type assertions fail.
	ErrServiceInvalid = errors.New("service invalid")
)
