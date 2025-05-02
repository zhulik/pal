package core

import (
	"errors"
)

var (
	ErrServiceNotFound   = errors.New("service not found")
	ErrServiceInitFailed = errors.New("service initialization failed")
	ErrServiceInvalid    = errors.New("service invalid")
)
