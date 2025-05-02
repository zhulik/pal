package pal

import (
	"context"
	"fmt"
	"reflect"

	"github.com/zhulik/pal/pkg/core"
)

type Service[I any, S any] struct {
	singleton bool
	runner    bool

	beforeInit core.LifecycleHook[S]

	instance I
}

// Name returns a name of the dependency derived from the interface.
func (f *Service[I, S]) Name() string {
	return elem[I]().String()
}

// Initialize creates a new instance of the Service, calls its Init method if it implements Initer.
func (f *Service[I, S]) Initialize(ctx context.Context) error {
	if !f.singleton {
		return nil
	}

	s, err := f.build(ctx)
	if err != nil {
		return err
	}

	f.instance = any(s).(I)

	return nil
}

func (f *Service[I, S]) Make() any {
	return new(S)
}

func (f *Service[I, S]) BeforeInit(hook core.LifecycleHook[S]) *Service[I, S] {
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
	iType := elem[I]()
	if iType.Kind() != reflect.Interface {
		return fmt.Errorf("%w: type parameter I (%v) must be an interface", core.ErrServiceInvalid, iType)
	}

	sType := elem[S]()
	if sType.Kind() != reflect.Struct {
		return fmt.Errorf("%w: type parameter S (%v) must be a struct", core.ErrServiceInvalid, sType)
	}

	if _, ok := any(new(S)).(I); !ok {
		return fmt.Errorf("%w: type %v does not implement interface %v", core.ErrServiceInvalid, sType, iType)
	}

	return nil
}

func (f *Service[I, S]) String() string {
	return fmt.Sprintf("%s[singleton=%v, runner=%v]", f.Name(), f.singleton, f.runner)
}

func (f *Service[I, S]) Instance(ctx context.Context) (any, error) {
	if f.singleton {
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

	if initer, ok := any(s).(core.Initer); ok {
		if err := initer.Init(ctx); err != nil {
			return nil, err
		}
	}
	return s, nil
}
