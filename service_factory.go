package pal

import (
	"context"
	"fmt"
	"reflect"
)

// ServiceFactory is a service that wraps a constant value which will copied and initialized on every invocation unlike
// ServiceConst, which is initialized only once and always return the same instance when invoked.
type ServiceFactory[T any] struct {
	P                 *Pal
	referenceInstance T

	hooks LifecycleHooks[T]
}

func (c *ServiceFactory[T]) Dependencies() []ServiceDef {
	return nil
}

func (c *ServiceFactory[T]) Run(_ context.Context) error {
	return nil
}

func (c *ServiceFactory[T]) Init(_ context.Context) error {
	return nil
}

func (c *ServiceFactory[T]) HealthCheck(_ context.Context) error {
	return nil
}

func (c *ServiceFactory[T]) Shutdown(_ context.Context) error {
	return nil
}

func (c *ServiceFactory[T]) Make() any {
	return c.referenceInstance
}

func (c *ServiceFactory[T]) Instance(ctx context.Context) (any, error) {
	logger := c.P.logger.With("service", c.Name())

	logger.Debug("Creating an instance")

	instance, err := c.copyInstance()
	if err != nil {
		return nil, err
	}

	err = c.P.InjectInto(ctx, instance)
	if err != nil {
		return nil, err
	}

	if c.hooks.Init != nil {
		logger.Debug("Calling BeforeInit hook")
		err := c.hooks.Init(ctx, instance.(T))
		if err != nil {
			c.P.logger.Error("BeforeInit hook failed", "error", err)
			return nil, err
		}
	}

	if initer, ok := instance.(Initer); ok {
		logger.Debug("Calling Init method")
		if err := initer.Init(ctx); err != nil {
			logger.Error("Init failed", "error", err)
			return nil, err
		}
	}
	return instance, nil
}

func (c *ServiceFactory[T]) copyInstance() (any, error) {
	refValue := reflect.ValueOf(c.referenceInstance)
	if !refValue.IsValid() {
		return nil, fmt.Errorf("invalid reference instance")
	}

	var newInstance reflect.Value
	if refValue.Kind() == reflect.Ptr {
		elemType := refValue.Elem().Type()
		newPtr := reflect.New(elemType)
		newPtr.Elem().Set(refValue.Elem())
		newInstance = newPtr
	} else {
		newInstance = reflect.New(refValue.Type()).Elem()
		newInstance.Set(refValue)
	}

	instance := newInstance.Interface()
	return instance, nil
}

func (c *ServiceFactory[T]) RunConfig() *RunConfig {
	return nil
}

func (c *ServiceFactory[T]) Name() string {
	return elem[T]().String()
}

func (c *ServiceFactory[T]) BeforeInit(hook LifecycleHook[T]) *ServiceFactory[T] {
	c.hooks.Init = hook
	return c
}
