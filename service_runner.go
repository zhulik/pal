package pal

import (
	"context"
	"fmt"
	"math/rand"
)

const idCharset = "abcdefghijklmnopqrstuvwxyz0123456789"

type ServiceRunner struct {
	P *Pal
	ServiceTyped[any]
	fn func(ctx context.Context) error
}

func (c *ServiceRunner) RunConfig() *RunConfig {
	return defaultRunConfig
}

func (c *ServiceRunner) Run(ctx context.Context) error {
	return runService(ctx, c.Name(), c.fn, c.P)
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
