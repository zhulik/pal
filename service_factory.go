package pal

import (
	"context"
	"reflect"
)

type serviceFactory[I any, S any] struct {
	singleton bool
	runner    bool
}

// Name returns a name of the dependency derived from the interface.
func (d serviceFactory[I, S]) Name() string {
	return reflect.TypeOf((*I)(nil)).Elem().String()
}

// Initialize creates a new instance of the service, calls its Init method if it implements Initer.
func (d serviceFactory[I, S]) Initialize(ctx context.Context) (any, error) {
	pal := ctx.Value(CtxValue).(*Pal)
	s := Inject[S](ctx, pal)

	if initer, ok := any(s).(Initer); ok {
		if err := initer.Init(ctx); err != nil {
			var i I
			return i, err
		}
	}

	return any(s).(I), nil
}

func (d serviceFactory[I, S]) Make() any {
	return *new(S)
}

func (d serviceFactory[I, S]) IsSingleton() bool {
	return d.singleton
}

func (d serviceFactory[I, S]) IsRunner() bool {
	return d.runner
}
