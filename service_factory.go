package pal

import (
	"context"
	"fmt"
)

type ServiceFactory[I any, S any] struct {
	P          *Pal
	beforeInit LifecycleHook[S]
}

func (c *ServiceFactory[I, S]) Run(_ context.Context) error {
	return nil
}

func (c *ServiceFactory[I, S]) Init(_ context.Context) error {
	return nil
}

func (c *ServiceFactory[I, S]) HealthCheck(_ context.Context) error {
	return nil
}

func (c *ServiceFactory[I, S]) Shutdown(_ context.Context) error {
	return nil
}

func (c *ServiceFactory[I, S]) Make() any {
	return empty[S]()
}

func (c *ServiceFactory[I, S]) Instance(ctx context.Context) (any, error) {
	return buildService[S](ctx, c.beforeInit, c.P, c.P.logger.With("service", c.Name()))
}

func (c *ServiceFactory[I, S]) Name() string {
	return elem[I]().String()
}

func (c *ServiceFactory[I, S]) Validate(ctx context.Context) error {
	return validateService[I, S](ctx)
}

func validateService[I any, S any](_ context.Context) error {
	iType := elem[I]()

	sType := elem[S]()

	if _, ok := any(new(S)).(I); !ok {
		return fmt.Errorf("%w: type %v does not implement interface %v", ErrServiceInvalid, sType, iType)
	}

	return nil
}

func (c *ServiceFactory[I, S]) BeforeInit(hook LifecycleHook[S]) *ServiceFactory[I, S] {
	c.beforeInit = hook
	return c
}
