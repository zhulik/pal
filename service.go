package pal

import (
	"context"
	"fmt"
)

type Service[I any, S any] struct {
	singleton bool
	runner    bool

	beforeInit LifecycleHook[S]

	instance I
}

// Name returns a name of the dependency derived from the interface.
func (f *Service[I, S]) Name() string {
	return elem[I]().String()
}

// Init creates a new instance of the Service, calls its Init method if it implements Initer.
func (f *Service[I, S]) Init(ctx context.Context) error {
	if !f.singleton {
		return nil
	}

	if !isNil(f.instance) {
		return nil
	}

	s, err := f.build(ctx)
	if err != nil {
		return err
	}

	f.instance = any(s).(I)

	return nil
}

func (f *Service[I, S]) HealthCheck(ctx context.Context) error {
	if !isNil(f.instance) {
		if h, ok := any(f.instance).(HealthChecker); ok {
			return h.HealthCheck(ctx)
		}
	}
	return nil
}

func (f *Service[I, S]) Shutdown(ctx context.Context) error {
	if !isNil(f.instance) {
		if h, ok := any(f.instance).(Shutdowner); ok {
			return h.Shutdown(ctx)
		}
	}
	return nil
}

func (f *Service[I, S]) Make() any {
	if !isNil(f.instance) {
		return nil
	}
	return new(S)
}

func (f *Service[I, S]) BeforeInit(hook LifecycleHook[S]) *Service[I, S] {
	f.beforeInit = hook
	return f
}

func (f *Service[I, S]) IsSingleton() bool {
	return f.singleton
}

func (f *Service[I, S]) IsRunner() bool {
	return f.runner
}

func (f *Service[I, S]) Validate(_ context.Context) error {
	if !isNil(f.instance) {
		return nil
	}
	iType := elem[I]()

	sType := elem[S]()

	if _, ok := any(new(S)).(I); !ok {
		return fmt.Errorf("%w: type %v does not implement interface %v", ErrServiceInvalid, sType, iType)
	}

	return nil
}

func (f *Service[I, S]) String() string {
	return fmt.Sprintf("%s[singleton=%v, runner=%v]", f.Name(), f.singleton, f.runner)
}

func (f *Service[I, S]) Instance(ctx context.Context) (any, error) {
	if f.singleton {
		if isNil(f.instance) {
			return nil, fmt.Errorf("%w: singleton service %s has not been initialized", ErrServiceInvalid, f.Name())
		}
		return f.instance, nil
	}

	return f.build(ctx)
}

func (f *Service[I, S]) build(ctx context.Context) (*S, error) {
	s, err := Inject[S](ctx, FromContext(ctx))
	if err != nil {
		return nil, err
	}

	if f.beforeInit != nil {
		err = f.beforeInit(ctx, s)
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
