package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhulik/pal"
	"github.com/zhulik/pal/pkg/core"
)

// TestService_Name tests the Name method of the service struct
func TestService_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns correct name for interface type", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestInterface, TestStruct]()

		assert.Equal(t, "pal_test.TestInterface", service.Name())
	})
}

// TestService_IsSingleton tests the IsSingleton method of the service struct
func TestService_IsSingleton(t *testing.T) {
	t.Parallel()

	t.Run("returns true for singleton service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestInterface, TestStruct]()

		assert.True(t, service.IsSingleton())
	})

	t.Run("returns false for factory service", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory[TestInterface, TestStruct]()

		assert.False(t, service.IsSingleton())
	})
}

// TestService_IsRunner tests the IsRunner method of the service struct
func TestService_IsRunner(t *testing.T) {
	t.Parallel()

	t.Run("returns true for runner service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[RunnerInterface, RunnerStruct]()

		assert.True(t, service.IsRunner())
	})

	t.Run("returns false for non-runner service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestInterface, TestStruct]()

		assert.False(t, service.IsRunner())
	})
}

// TestService_Make tests the Make method of the service struct
func TestService_Make(t *testing.T) {
	t.Parallel()

	t.Run("returns empty instance of service type", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestInterface, TestStruct]()

		assert.Implements(t, (*TestInterface)(nil), service.Make())
	})
}

// TestService_Validate tests the Validate method of the service struct
func TestService_Validate(t *testing.T) {
	t.Parallel()

	t.Run("validates service successfully", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestInterface, TestStruct]()

		assert.NoError(t, service.Validate(t.Context()))
	})

	t.Run("returns error when I is not an interface", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestStruct, TestStruct]()

		err := service.Validate(t.Context())

		assert.Error(t, err)
		assert.ErrorIs(t, err, core.ErrServiceInvalid)
	})

	t.Run("returns error when S is not a struct", func(t *testing.T) {
		t.Parallel()

		type NotAStruct string
		service := pal.Provide[TestInterface, NotAStruct]()

		err := service.Validate(t.Context())

		assert.Error(t, err)
		assert.ErrorIs(t, err, core.ErrServiceInvalid)
	})

	t.Run("returns error when S does not implement I", func(t *testing.T) {
		t.Parallel()

		type NonImplementingStruct struct{}
		service := pal.Provide[TestInterface, NonImplementingStruct]()

		err := service.Validate(t.Context())

		assert.Error(t, err)
		assert.ErrorIs(t, err, core.ErrServiceInvalid)
	})
}

// TestService_Instance tests the Instance method of the service struct
func TestService_Instance(t *testing.T) {
	t.Parallel()

	t.Run("returns instance for singleton service", func(t *testing.T) {
		t.Parallel()

		service := pal.Provide[TestInterface, TestStruct]()

		p := pal.New(service)

		err := p.Init(t.Context())
		assert.NoError(t, err)

		instance1, err := service.Instance(t.Context())

		assert.NoError(t, err)
		assert.NotNil(t, instance1)

		instance2, err := service.Instance(t.Context())

		assert.NoError(t, err)
		assert.NotNil(t, instance1)
		assert.Same(t, instance1, instance2)
	})

	t.Run("returns new instance for factory service", func(t *testing.T) {
		t.Parallel()

		service := pal.ProvideFactory[TestInterface, TestStruct]()
		p := pal.New(service)
		ctx := context.WithValue(t.Context(), pal.CtxValue, p)

		// First call to Instance should create a new instance
		instance, err := service.Instance(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, instance)

		// Second call should create another new instance
		instance2, err := service.Instance(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, instance2)
		assert.NotSame(t, instance, instance2)
	})
}
