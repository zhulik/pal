package pal

import (
	"context"
	"fmt"
	"math/rand"
)

const idCharset = "abcdefghijklmnopqrstuvwxyz0123456789"

type ServiceRunner struct {
	fn func(ctx context.Context) error
}

func (c *ServiceRunner) Dependencies() []ServiceDef {
	return nil
}

func (c *ServiceRunner) Run(ctx context.Context) error {
	return c.fn(ctx)
}

func (c *ServiceRunner) Init(_ context.Context) error {
	return nil
}

func (c *ServiceRunner) HealthCheck(_ context.Context) error {
	return nil
}

func (c *ServiceRunner) Shutdown(_ context.Context) error {
	return nil
}

func (c *ServiceRunner) Make() any {
	return nil
}

func (c *ServiceRunner) Instance(_ context.Context, _ ...any) (any, error) {
	return nil, nil
}

func (c *ServiceRunner) Name() string {
	return fmt.Sprintf("function-runner-%s", randomID())
}

func (c *ServiceRunner) RunConfig() *RunConfig {
	return nil
}

func randomID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = idCharset[rand.Intn(len(idCharset))] // nolint:gosec
	}

	return string(b)
}
