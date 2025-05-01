package pal

import (
	"context"
	"fmt"
	"reflect"
)

type serviceFactory[I any, S any] struct {
	singleton bool
	runner    bool
}

// Name returns a name of the dependency derived from the interface.
func (f serviceFactory[I, S]) Name() string {
	return elem[I]().String()
}

// Initialize creates a new instance of the service, calls its Init method if it implements Initer.
func (f serviceFactory[I, S]) Initialize(ctx context.Context) (any, error) {
	s, err := Inject[S](ctx, FromContext(ctx))
	if err != nil {
		return nil, err
	}

	if initer, ok := any(s).(Initer); ok {
		if err := initer.Init(ctx); err != nil {
			return empty[I](), err
		}
	}

	return s, nil
}

func (f serviceFactory[I, S]) Make() any {
	return empty[S]()
}

func (f serviceFactory[I, S]) IsSingleton() bool {
	return f.singleton
}

func (f serviceFactory[I, S]) IsRunner() bool {
	return f.runner
}

func (f serviceFactory[I, S]) Validate(_ context.Context) error {
	iType := elem[I]()
	if iType.Kind() != reflect.Interface {
		return fmt.Errorf("%w: type parameter I (%v) must be an interface", ErrServiceInvalid, iType)
	}

	sType := elem[S]()
	if sType.Kind() != reflect.Struct {
		return fmt.Errorf("%w: type parameter S (%v) must be a struct", ErrServiceInvalid, sType)
	}

	if !sType.Implements(iType) {
		return fmt.Errorf("%w: type %v does not implement interface %v", ErrServiceInvalid, sType, iType)
	}

	return nil
}

func (f serviceFactory[I, S]) String() string {
	return fmt.Sprintf("%s[singleton=%v, runner=%v]", f.Name(), f.singleton, f.runner)
}
