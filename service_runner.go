package pal

import (
	"context"
	"fmt"
	"math/rand"
)

const idCharset = "abcdefghijklmnopqrstuvwxyz0123456789"

type ServiceRunner struct {
	ServiceTyped[any]
	fn func(ctx context.Context) error
}

func (c *ServiceRunner) Run(ctx context.Context) error {
	return c.fn(ctx)
}

func (c *ServiceRunner) Instance(_ context.Context, _ ...any) (any, error) {
	return nil, nil
}

func (c *ServiceRunner) Name() string {
	return fmt.Sprintf("$function-runner-%s", randomID())
}

func randomID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = idCharset[rand.Intn(len(idCharset))] // nolint:gosec
	}

	return string(b)
}
