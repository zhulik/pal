package core

import "context"

// Pinger is an interface for the pinger service.
type Pinger interface {
	Ping(ctx context.Context) error
}
