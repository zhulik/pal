package pal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/zhulik/pal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhulik/pal/pkg/core"
)

// MockService implements the core.Service interface for testing
type MockService struct {
	mock.Mock
	name        string
	isSingleton bool
	isRunner    bool
	instance    *MockInstance
}

func NewMockService(name string, isSingleton bool, isRunner bool, instance ...*MockInstance) *MockService {
	var instancePtr *MockInstance
	if len(instance) > 0 {
		instancePtr = instance[0]
	}
	return &MockService{
		name:        name,
		isSingleton: isSingleton,
		isRunner:    isRunner,
		instance:    instancePtr,
	}
}

func (m *MockService) Make() any {
	if m.instance != nil {
		return m.instance
	}
	args := m.Called()
	return args.Get(0)
}

func (m *MockService) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockService) Instance(ctx context.Context) (any, error) {
	if m.instance != nil {
		return m.instance, nil
	}
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
}

func (m *MockService) Name() string {
	return m.name
}

func (m *MockService) IsSingleton() bool {
	return m.isSingleton
}

func (m *MockService) IsRunner() bool {
	return m.isRunner
}

func (m *MockService) Validate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockInstance implements various optional interfaces for testing
type MockInstance struct {
	mock.Mock
}

func (m *MockInstance) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockInstance) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockInstance) Run(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestContainer_New tests the New function for Container
func TestContainer_New(t *testing.T) {
	t.Parallel()

	t.Run("creates a new Container with services", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer(
			NewMockService("service1", true, false),
			NewMockService("service2", true, true),
		)

		assert.NotNil(t, c)
	})

	t.Run("creates a new Container with empty services", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer()

		assert.NotNil(t, c)
		// We can verify it works with nil services by checking that Services() returns empty
		assert.Empty(t, c.Services())
	})
}

// TestContainer_Validate tests the Validate method of Container
func TestContainer_Validate(t *testing.T) {
	t.Parallel()

	t.Run("validates all services successfully", func(t *testing.T) {
		t.Parallel()

		instance1 := newMockInstance(t)
		instance2 := newMockInstance(t)

		service1 := NewMockService("service1", true, false, instance1)
		service2 := NewMockService("service2", true, true, instance2)

		service1.On("Validate", t.Context()).Return(nil)
		service2.On("Validate", t.Context()).Return(nil)

		c := pal.NewContainer(service1, service2)

		err := c.Validate(t.Context())

		assert.NoError(t, err)
	})

	t.Run("returns error when service validation fails", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockService("service1", true, false)
		service2 := NewMockService("service2", true, true)

		service1.On("Validate", t.Context()).Return(nil)
		service2.On("Validate", t.Context()).Return(errTest)

		c := pal.NewContainer(service1, service2)

		err := c.Validate(t.Context())

		assert.ErrorIs(t, err, errTest)
	})
}

func newMockInstance(t *testing.T) *MockInstance {
	t.Helper()

	m := &MockInstance{}
	t.Cleanup(func() {
		m.AssertExpectations(t)
	})

	return m
}

// TestContainer_Init tests the Init method of Container
func TestContainer_Init(t *testing.T) {
	t.Parallel()

	t.Run("initializes singleton services successfully", func(t *testing.T) {
		t.Parallel()

		instance1 := newMockInstance(t)
		instance2 := newMockInstance(t)
		instance3 := newMockInstance(t)

		service1 := NewMockService("service1", true, false, instance1)
		service2 := NewMockService("service2", true, true, instance2)
		service3 := NewMockService("service3", true, false, instance3)

		service1.On("Initialize", t.Context()).Return(nil)
		service2.On("Initialize", t.Context()).Return(nil)
		service3.On("Initialize", t.Context()).Return(nil)

		c := pal.NewContainer(service1, service2, service3)

		err := c.Init(t.Context())

		assert.NoError(t, err)
	})

	t.Run("returns error when service initialization fails", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockService("service1", true, false)
		service2 := NewMockService("service2", true, false)

		service1.On("Make").Return(nil)
		service2.On("Make").Return(nil)

		service1.On("Initialize", t.Context()).Return(nil)
		service2.On("Initialize", t.Context()).Return(errTest)

		c := pal.NewContainer(service1, service2)

		err := c.Init(t.Context())

		assert.ErrorIs(t, err, errTest)
	})
}

// TestContainer_Invoke tests the Invoke method of Container
func TestContainer_Invoke(t *testing.T) {
	t.Parallel()

	t.Run("invokes service successfully", func(t *testing.T) {
		t.Parallel()

		expectedInstance := newMockInstance(t)

		service := NewMockService("service1", true, false, expectedInstance)
		service.On("Initialize", t.Context()).Return(nil)

		c := pal.NewContainer(service)
		require.NoError(t, c.Init(t.Context()))

		instance, err := c.Invoke(t.Context(), "service1")

		assert.NoError(t, err)
		assert.Equal(t, expectedInstance, instance)
	})

	t.Run("returns error when service not found", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer()

		_, err := c.Invoke(t.Context(), "nonexistent")

		assert.ErrorIs(t, err, core.ErrServiceNotFound)
	})

	t.Run("returns error when service instance creation fails", func(t *testing.T) {
		t.Parallel()

		service := NewMockService("service1", true, false)
		service.On("Make").Return(nil)
		service.On("Initialize", t.Context()).Return(nil)
		service.On("Instance", t.Context()).Return(nil, errTest)

		c := pal.NewContainer(service)
		require.NoError(t, c.Init(t.Context()))

		_, err := c.Invoke(t.Context(), "service1")

		assert.ErrorIs(t, err, core.ErrServiceInitFailed)
	})
}

// TestContainer_Shutdown tests the Shutdown method of Container
func TestContainer_Shutdown(t *testing.T) {
	t.Parallel()

	t.Run("shuts down all singleton services successfully", func(t *testing.T) {
		t.Parallel()
		instance1 := newMockInstance(t)
		instance2 := newMockInstance(t)
		instance3 := newMockInstance(t)

		service1 := NewMockService("service1", true, false, instance1)
		service2 := NewMockService("service2", true, true, instance2)
		service3 := NewMockService("service3", true, false, instance3)

		service1.On("Initialize", t.Context()).Return(nil)
		service2.On("Initialize", t.Context()).Return(nil)
		service3.On("Initialize", t.Context()).Return(nil)

		instance1.On("Shutdown", t.Context()).Return(nil)
		instance2.On("Shutdown", t.Context()).Return(nil)
		instance3.On("Shutdown", t.Context()).Return(nil)

		c := pal.NewContainer(service1, service2, service3)
		require.NoError(t, c.Init(t.Context()))

		err := c.Shutdown(t.Context())

		assert.NoError(t, err)
	})

	t.Run("returns error when service shutdown fails", func(t *testing.T) {
		t.Parallel()

		instance := newMockInstance(t)
		instance.On("Shutdown", t.Context()).Return(errTest)

		service := NewMockService("service1", true, false, instance)
		service.On("Initialize", t.Context()).Return(nil)

		c := pal.NewContainer(service)
		require.NoError(t, c.Init(t.Context()))

		err := c.Shutdown(t.Context())

		assert.ErrorIs(t, err, errTest)
	})
}

// TestContainer_HealthCheck tests the HealthCheck method of Container
func TestContainer_HealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("health checks all singleton services successfully", func(t *testing.T) {
		t.Parallel()

		instance1 := newMockInstance(t)
		instance1.On("HealthCheck", t.Context()).Return(nil)
		instance2 := newMockInstance(t)
		instance2.On("HealthCheck", t.Context()).Return(nil)
		instance3 := newMockInstance(t)
		instance3.On("HealthCheck", t.Context()).Return(nil)

		service1 := NewMockService("service1", true, false, instance1)
		service2 := NewMockService("service2", true, true, instance2)
		service3 := NewMockService("service3", true, false, instance3)

		service1.On("Initialize", t.Context()).Return(nil)
		service2.On("Initialize", t.Context()).Return(nil)
		service3.On("Initialize", t.Context()).Return(nil)

		c := pal.NewContainer(service1, service2, service3)
		require.NoError(t, c.Init(t.Context()))

		err := c.HealthCheck(t.Context())

		assert.NoError(t, err)
	})

	t.Run("returns error when service health check fails", func(t *testing.T) {
		t.Parallel()

		instance := newMockInstance(t)
		instance.On("HealthCheck", t.Context()).Return(errTest)

		service := NewMockService("service1", true, false, instance)
		service.On("Initialize", t.Context()).Return(nil)

		c := pal.NewContainer(service)
		require.NoError(t, c.Init(t.Context()))

		err := c.HealthCheck(t.Context())

		assert.ErrorIs(t, err, errTest)
	})
}

// TestContainer_Services tests the Services method of Container
func TestContainer_Services(t *testing.T) {
	t.Parallel()

	t.Run("returns all services", func(t *testing.T) {
		t.Parallel()

		instance1 := newMockInstance(t)
		instance2 := newMockInstance(t)

		service1 := NewMockService("service1", true, false, instance1)
		service2 := NewMockService("service2", true, true, instance2)

		service1.On("Initialize", t.Context()).Return(nil)
		service2.On("Initialize", t.Context()).Return(nil)

		c := pal.NewContainer(service1, service2)
		require.NoError(t, c.Init(t.Context()))

		result := c.Services()

		assert.Len(t, result, 2)
		assert.Contains(t, result, service1)
		assert.Contains(t, result, service2)
	})

	t.Run("returns empty slice for empty container", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer()

		result := c.Services()

		assert.Empty(t, result)
	})
}

// TestContainer_SetLogger tests the SetLogger method of Container
func TestContainer_SetLogger(t *testing.T) {
	t.Parallel()

	t.Run("sets logger function", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer()

		var logCalled bool
		var logMessage string
		var logArgs []any

		logger := func(fmt string, args ...any) {
			logCalled = true
			logMessage = fmt
			logArgs = args
		}

		c.SetLogger(logger)

		// Add a service to trigger logging
		instance := newMockInstance(t)
		service := NewMockService("service1", true, false, instance)
		service.On("Initialize", t.Context()).Return(nil)
		instance.On("Shutdown", t.Context()).Return(nil)

		c = pal.NewContainer(service)
		c.SetLogger(logger)

		require.NoError(t, c.Init(t.Context()))
		require.NoError(t, c.Shutdown(t.Context()))

		assert.True(t, logCalled)
		assert.Contains(t, logMessage, "%s")
		assert.Equal(t, "service1", logArgs[0])
	})
}
