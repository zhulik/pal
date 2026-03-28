package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal"
)

type factoryMakeInterface interface {
	marker()
}

// A non–zero-sized struct so reflect.New produces distinct pointers each time.
type factoryMakeConcrete struct{ _ byte }

func (*factoryMakeConcrete) marker() {}

func TestServiceFactory_Make_returnsDistinctAllocatedImplementations(t *testing.T) {
	t.Parallel()

	s := pal.ProvideNamedFactory0[factoryMakeInterface](
		"factoryMake",
		func(_ context.Context) (*factoryMakeConcrete, error) {
			return &factoryMakeConcrete{}, nil
		},
	)

	a := s.Make()
	b := s.Make()
	ai, ok := a.(factoryMakeInterface)
	require.True(t, ok)
	require.NotNil(t, ai)
	bi, ok := b.(factoryMakeInterface)
	require.True(t, ok)
	assert.NotSame(t, ai, bi)
}
