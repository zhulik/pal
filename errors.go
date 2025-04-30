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
