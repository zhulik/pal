package pal

import (
	"context"
	"fmt"
)

type FactoryService[I any, S any] struct {
	beforeInit LifecycleHook[S]
}

func (c *FactoryService[I, S]) Init(_ context.Context) error {
	return nil
}

func (c *FactoryService[I, S]) HealthCheck(_ context.Context) error {
	return nil
}

func (c *FactoryService[I, S]) Shutdown(_ context.Context) error {
	return nil
}

func (c *FactoryService[I, S]) Make() any {
	return empty[S]()
}

func (c *FactoryService[I, S]) Instance(ctx context.Context) (any, error) {
	return buildInstance[S](ctx, c.beforeInit)
}

func (c *FactoryService[I, S]) Name() string {
	return elem[I]().String()
}

func (c *FactoryService[I, S]) IsSingleton() bool {
	return false
}

func (c *FactoryService[I, S]) IsRunner() bool {
	return false
}

func (c *FactoryService[I, S]) Validate(ctx context.Context) error {
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

func (c *FactoryService[I, S]) BeforeInit(hook LifecycleHook[S]) *FactoryService[I, S] {
	c.beforeInit = hook
	return c
}

func buildInstance[S any](ctx context.Context, beforeInit LifecycleHook[S]) (*S, error) {
	s, err := Inject[S](ctx, FromContext(ctx))
	if err != nil {
		return nil, err
	}

	if beforeInit != nil {
		err = beforeInit(ctx, s)
		if err != nil {
			return nil, err
		}
	}

	if initer, ok := any(s).(Initer); ok {
		if err := initer.Init(ctx); err != nil {
			return nil, err
		}
	}
	return s, nil
}
