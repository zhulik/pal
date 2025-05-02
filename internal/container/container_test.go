package container_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhulik/pal/internal/container"
	"github.com/zhulik/pal/pkg/core"
)

// MockService implements the core.Service interface for testing
type MockService struct {
	name           string
	isSingleton    bool
	isRunner       bool
	validateErr    error
	initializeErr  error
	instanceErr    error
	instance       any
	initialized    bool
	validateCalled bool
	initCalled     bool
	instanceCalled bool
}

func NewMockService(name string, isSingleton bool, isRunner bool) *MockService {
	return &MockService{
		name:        name,
		isSingleton: isSingleton,
		isRunner:    isRunner,
		instance:    &MockInstance{name: name},
	}
}

func (m *MockService) Make() any {
	return m.instance
}

func (m *MockService) Initialize(_ context.Context) error {
	m.initCalled = true
	m.initialized = true
	return m.initializeErr
}

func (m *MockService) Instance(_ context.Context) (any, error) {
	m.instanceCalled = true
	if m.instanceErr != nil {
		return nil, m.instanceErr
	}
	return m.instance, nil
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

func (m *MockService) Validate(_ context.Context) error {
	m.validateCalled = true
	return m.validateErr
}

// MockInstance implements various optional interfaces for testing
type MockInstance struct {
	name              string
	shutdownCalled    bool
	shutdownErr       error
	healthCheckCalled bool
	healthCheckErr    error
	runCalled         bool
	runErr            error
}

func (m *MockInstance) Shutdown(_ context.Context) error {
	m.shutdownCalled = true
	return m.shutdownErr
}

func (m *MockInstance) HealthCheck(_ context.Context) error {
	m.healthCheckCalled = true
	return m.healthCheckErr
}

func (m *MockInstance) Run(_ context.Context) error {
	m.runCalled = true
	return m.runErr
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("creates a new Container with services", func(t *testing.T) {
		t.Parallel()

		services := map[string]core.Service{
			"service1": NewMockService("service1", true, false),
			"service2": NewMockService("service2", false, true),
		}

		c := container.New(services)

		assert.NotNil(t, c)
		// We can't directly access private fields, so we'll test functionality instead
	})

	t.Run("creates a new Container with nil services", func(t *testing.T) {
		t.Parallel()

		c := container.New(nil)

		assert.NotNil(t, c)
		// We can verify it works with nil services by checking that Services() returns empty
		assert.Empty(t, c.Services())
	})
}

// TestValidate tests the Validate method
func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("validates all services successfully", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockService("service1", true, false)
		service2 := NewMockService("service2", false, true)

		services := map[string]core.Service{
			"service1": service1,
			"service2": service2,
		}

		c := container.New(services)

		err := c.Validate(context.Background())

		assert.NoError(t, err)
		assert.True(t, service1.validateCalled)
		assert.True(t, service2.validateCalled)
	})

	t.Run("returns error when service validation fails", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockService("service1", true, false)
		service2 := NewMockService("service2", false, true)
		expectedErr := errors.New("validation error")
		service2.validateErr = expectedErr

		services := map[string]core.Service{
			"service1": service1,
			"service2": service2,
		}

		c := container.New(services)

		err := c.Validate(context.Background())

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.True(t, service1.validateCalled)
		assert.True(t, service2.validateCalled)
	})
}

// TestInit tests the Init method
func TestInit(t *testing.T) {
	t.Parallel()

	t.Run("initializes singleton services successfully", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockService("service1", true, false)
		service2 := NewMockService("service2", false, true)
		service3 := NewMockService("service3", true, false)

		services := map[string]core.Service{
			"service1": service1,
			"service2": service2,
			"service3": service3,
		}

		c := container.New(services)

		err := c.Init(context.Background())

		assert.NoError(t, err)
		assert.True(t, service1.initCalled)
		assert.False(t, service2.initCalled) // Not a singleton
		assert.True(t, service3.initCalled)
	})

	t.Run("returns error when service initialization fails", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockService("service1", true, false)
		service2 := NewMockService("service2", true, false)
		expectedErr := errors.New("init error")
		service2.initializeErr = expectedErr

		services := map[string]core.Service{
			"service1": service1,
			"service2": service2,
		}

		c := container.New(services)

		err := c.Init(context.Background())

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

// TestInvoke tests the Invoke method
func TestInvoke(t *testing.T) {
	t.Parallel()

	t.Run("invokes service successfully", func(t *testing.T) {
		t.Parallel()

		service := NewMockService("service1", true, false)
		expectedInstance := service.instance

		services := map[string]core.Service{
			"service1": service,
		}

		c := container.New(services)
		require.NoError(t, c.Init(context.Background()))

		instance, err := c.Invoke(context.Background(), "service1")

		assert.NoError(t, err)
		assert.Equal(t, expectedInstance, instance)
		assert.True(t, service.instanceCalled)
	})

	t.Run("returns error when service not found", func(t *testing.T) {
		t.Parallel()

		c := container.New(nil)

		_, err := c.Invoke(context.Background(), "nonexistent")

		assert.Error(t, err)
		assert.ErrorIs(t, err, core.ErrServiceNotFound)
	})

	t.Run("returns error when service instance creation fails", func(t *testing.T) {
		t.Parallel()

		service := NewMockService("service1", true, false)
		service.instanceErr = errors.New("instance error")

		services := map[string]core.Service{
			"service1": service,
		}

		c := container.New(services)
		require.NoError(t, c.Init(context.Background()))

		_, err := c.Invoke(context.Background(), "service1")

		assert.Error(t, err)
		assert.ErrorIs(t, err, core.ErrServiceInitFailed)
	})
}

// TestShutdown tests the Shutdown method
func TestShutdown(t *testing.T) {
	t.Parallel()

	t.Run("shuts down all singleton services successfully", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockService("service1", true, false)
		service2 := NewMockService("service2", false, true) // Not a singleton
		service3 := NewMockService("service3", true, false)

		services := map[string]core.Service{
			"service1": service1,
			"service2": service2,
			"service3": service3,
		}

		c := container.New(services)
		require.NoError(t, c.Init(context.Background()))

		err := c.Shutdown(context.Background())

		assert.NoError(t, err)

		// Check that Shutdown was called on the instances
		instance1 := service1.instance.(*MockInstance)
		instance2 := service2.instance.(*MockInstance)
		instance3 := service3.instance.(*MockInstance)

		assert.True(t, instance1.shutdownCalled)
		assert.False(t, instance2.shutdownCalled) // Not a singleton
		assert.True(t, instance3.shutdownCalled)
	})

	t.Run("returns error when service shutdown fails", func(t *testing.T) {
		t.Parallel()

		service := NewMockService("service1", true, false)
		instance := service.instance.(*MockInstance)
		expectedErr := errors.New("shutdown error")
		instance.shutdownErr = expectedErr

		services := map[string]core.Service{
			"service1": service,
		}

		c := container.New(services)
		require.NoError(t, c.Init(context.Background()))

		err := c.Shutdown(context.Background())

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.True(t, instance.shutdownCalled)
	})
}

// TestHealthCheck tests the HealthCheck method
func TestHealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("health checks all singleton services successfully", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockService("service1", true, false)
		service2 := NewMockService("service2", false, true) // Not a singleton
		service3 := NewMockService("service3", true, false)

		services := map[string]core.Service{
			"service1": service1,
			"service2": service2,
			"service3": service3,
		}

		c := container.New(services)
		require.NoError(t, c.Init(context.Background()))

		err := c.HealthCheck(context.Background())

		assert.NoError(t, err)

		// Check that HealthCheck was called on the instances
		instance1 := service1.instance.(*MockInstance)
		instance2 := service2.instance.(*MockInstance)
		instance3 := service3.instance.(*MockInstance)

		assert.True(t, instance1.healthCheckCalled)
		assert.False(t, instance2.healthCheckCalled) // Not a singleton
		assert.True(t, instance3.healthCheckCalled)
	})

	t.Run("returns error when service health check fails", func(t *testing.T) {
		t.Parallel()

		service := NewMockService("service1", true, false)
		instance := service.instance.(*MockInstance)
		expectedErr := errors.New("health check error")
		instance.healthCheckErr = expectedErr

		services := map[string]core.Service{
			"service1": service,
		}

		c := container.New(services)
		require.NoError(t, c.Init(context.Background()))

		err := c.HealthCheck(context.Background())

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.True(t, instance.healthCheckCalled)
	})
}

// TestServices tests the Services method
func TestServices(t *testing.T) {
	t.Parallel()

	t.Run("returns all services", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockService("service1", true, false)
		service2 := NewMockService("service2", false, true)

		services := map[string]core.Service{
			"service1": service1,
			"service2": service2,
		}

		c := container.New(services)
		require.NoError(t, c.Init(context.Background()))

		result := c.Services()

		assert.Len(t, result, 2)
		assert.Contains(t, result, service1)
		assert.Contains(t, result, service2)
	})

	t.Run("returns empty slice for empty container", func(t *testing.T) {
		t.Parallel()

		c := container.New(nil)

		result := c.Services()

		assert.Empty(t, result)
	})
}

// TestRunners tests the Runners method
func TestRunners(t *testing.T) {
	t.Parallel()

	t.Run("returns all runner services", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockService("service1", true, false) // Not a runner
		service2 := NewMockService("service2", true, true)  // Runner
		service3 := NewMockService("service3", false, true) // Runner but not singleton

		services := map[string]core.Service{
			"service1": service1,
			"service2": service2,
			"service3": service3,
		}

		c := container.New(services)
		require.NoError(t, c.Init(context.Background()))

		runners := c.Runners(context.Background())

		assert.Len(t, runners, 2)
		assert.Contains(t, runners, "service2")
		assert.Contains(t, runners, "service3")
		assert.NotContains(t, runners, "service1")
	})

	t.Run("returns empty map when no runners", func(t *testing.T) {
		t.Parallel()

		service := NewMockService("service1", true, false) // Not a runner

		services := map[string]core.Service{
			"service1": service,
		}

		c := container.New(services)
		require.NoError(t, c.Init(context.Background()))

		runners := c.Runners(context.Background())

		assert.Empty(t, runners)
	})

	t.Run("skips runner with instance error", func(t *testing.T) {
		t.Parallel()

		service := NewMockService("service1", true, true)
		service.instanceErr = errors.New("instance error")

		services := map[string]core.Service{
			"service1": service,
		}

		c := container.New(services)
		require.NoError(t, c.Init(context.Background()))

		runners := c.Runners(context.Background())

		assert.Empty(t, runners)
	})
}
