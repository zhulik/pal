package pal_test

import (
	"testing"

	"github.com/zhulik/pal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewMockService(t *testing.T, name string) *MockServiceDef {
	t.Helper()
	mock := NewMockServiceDef(t)

	mock.EXPECT().Name().Return(name).Maybe()
	mock.EXPECT().Dependencies().Return(nil).Maybe()
	mock.EXPECT().Make().Return(nil).Maybe()
	mock.EXPECT().Arguments().Return(0).Maybe()
	mock.EXPECT().ShouldWaitForRunner().Return(nil).Maybe()

	return mock
}

// mockLifecycleService is a ServiceDef that also drives Init/Shutdown/HealthCheck on the definition,
// matching Provide* wrappers after lifecycle was removed from ServiceDef itself.
type mockLifecycleService struct {
	*MockServiceDef
	*MockIniter
	*MockShutdowner
	*MockHealthChecker
}

func NewMockLifecycleService(t *testing.T, name string) *mockLifecycleService {
	t.Helper()
	return &mockLifecycleService{
		MockServiceDef:    NewMockService(t, name),
		MockIniter:        NewMockIniter(t),
		MockShutdowner:    NewMockShutdowner(t),
		MockHealthChecker: NewMockHealthChecker(t),
	}
}

type cycleServiceA struct {
	B *cycleServiceB
}

type cycleServiceB struct {
	C *cycleServiceC
}

type cycleServiceC struct {
	A *cycleServiceA
}

// TestContainer_New tests the New function for Container
func TestContainer_New(t *testing.T) {
	t.Parallel()

	t.Run("creates a new Container with services", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer(
			&pal.Pal{},
			NewMockService(t, "service1"),
			NewMockService(t, "service2"),
		)

		assert.NotNil(t, c)
	})

	t.Run("creates a new Container with empty services", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer(&pal.Pal{})

		assert.NotNil(t, c)
		// We can verify it works with nil services by checking that Services() returns empty
		assert.Empty(t, c.Services())
	})
}

// TestContainer_Init tests the Init method of Container
func TestContainer_Init(t *testing.T) {
	t.Parallel()

	t.Run("skips services without Init on the definition", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer(
			&pal.Pal{},
			NewMockService(t, "service1"),
			NewMockService(t, "service2"),
			NewMockService(t, "service3"),
		)

		err := c.Init(t.Context())

		assert.NoError(t, err)
	})

	t.Run("initializes services that implement Init on the definition", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockLifecycleService(t, "service1")
		service2 := NewMockLifecycleService(t, "service2")
		service3 := NewMockLifecycleService(t, "service3")

		service1.MockIniter.EXPECT().Init(t.Context()).Return(nil)
		service2.MockIniter.EXPECT().Init(t.Context()).Return(nil)
		service3.MockIniter.EXPECT().Init(t.Context()).Return(nil)

		c := pal.NewContainer(&pal.Pal{}, service1, service2, service3)

		err := c.Init(t.Context())

		assert.NoError(t, err)
	})

	t.Run("returns error when service initialization fails", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockLifecycleService(t, "service1")
		service2 := NewMockLifecycleService(t, "service2")

		service1.MockIniter.EXPECT().Init(t.Context()).Return(nil).Maybe() // Init order is not guaranteed
		service2.MockIniter.EXPECT().Init(t.Context()).Return(errTest).Once()

		c := pal.NewContainer(&pal.Pal{}, service1, service2)

		err := c.Init(t.Context())

		assert.ErrorIs(t, err, errTest)
	})

	t.Run("returns cycle with service names in error", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer(
			&pal.Pal{},
			pal.Provide(&cycleServiceA{}),
			pal.Provide(&cycleServiceB{}),
			pal.Provide(&cycleServiceC{}),
		)

		err := c.Init(t.Context())

		require.Error(t, err)
		assert.ErrorContains(t, err, "cycleServiceA")
		assert.ErrorContains(t, err, "cycleServiceB")
		assert.ErrorContains(t, err, "cycleServiceC")
	})
}

// TestContainer_Invoke tests the Invoke method of Container
func TestContainer_Invoke(t *testing.T) {
	t.Parallel()

	t.Run("invokes service successfully", func(t *testing.T) {
		t.Parallel()

		expectedInstance := struct{}{}

		service := NewMockService(t, "service1")
		service.EXPECT().Instance(t.Context()).Return(expectedInstance, nil)

		c := pal.NewContainer(&pal.Pal{}, service)
		require.NoError(t, c.Init(t.Context()))

		instance, err := c.Invoke(t.Context(), "service1")

		assert.NoError(t, err)
		assert.Exactly(t, expectedInstance, instance)
	})

	t.Run("returns error when service not found", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer(&pal.Pal{})

		_, err := c.Invoke(t.Context(), "nonexistent")

		assert.ErrorIs(t, err, pal.ErrServiceNotFound)
	})

	t.Run("returns error when service instance creation fails", func(t *testing.T) {
		t.Parallel()

		service := NewMockService(t, "service1")
		service.EXPECT().Instance(t.Context()).Return(nil, errTest)

		c := pal.NewContainer(&pal.Pal{}, service)
		require.NoError(t, c.Init(t.Context()))

		_, err := c.Invoke(t.Context(), "service1")

		assert.ErrorIs(t, err, pal.ErrServiceInitFailed)
	})
}

// TestContainer_Shutdown tests the Shutdown method of Container
func TestContainer_Shutdown(t *testing.T) {
	t.Parallel()

	t.Run("skips services without Shutdown on the definition", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer(
			&pal.Pal{},
			NewMockService(t, "service1"),
			NewMockService(t, "service2"),
		)
		require.NoError(t, c.Init(t.Context()))

		err := c.Shutdown(t.Context())

		assert.NoError(t, err)
	})

	t.Run("shuts down services that implement Shutdown on the definition", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockLifecycleService(t, "service1")
		service2 := NewMockLifecycleService(t, "service2")
		service3 := NewMockLifecycleService(t, "service3")

		service1.MockIniter.EXPECT().Init(t.Context()).Return(nil)
		service2.MockIniter.EXPECT().Init(t.Context()).Return(nil)
		service3.MockIniter.EXPECT().Init(t.Context()).Return(nil)

		service1.MockShutdowner.EXPECT().Shutdown(t.Context()).Return(nil)
		service2.MockShutdowner.EXPECT().Shutdown(t.Context()).Return(nil)
		service3.MockShutdowner.EXPECT().Shutdown(t.Context()).Return(nil)

		c := pal.NewContainer(&pal.Pal{}, service1, service2, service3)
		require.NoError(t, c.Init(t.Context()))

		err := c.Shutdown(t.Context())

		assert.NoError(t, err)
	})

	t.Run("returns error when service shutdown fails", func(t *testing.T) {
		t.Parallel()

		service := NewMockLifecycleService(t, "service1")
		service.MockIniter.EXPECT().Init(t.Context()).Return(nil)
		service.MockShutdowner.EXPECT().Shutdown(t.Context()).Return(errTest)

		c := pal.NewContainer(&pal.Pal{}, service)
		require.NoError(t, c.Init(t.Context()))

		err := c.Shutdown(t.Context())

		assert.ErrorIs(t, err, errTest)
	})
}

// TestContainer_HealthCheck tests the HealthCheck method of Container
func TestContainer_HealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("skips services without HealthCheck on the definition", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer(
			&pal.Pal{},
			NewMockService(t, "service1"),
			NewMockService(t, "service2"),
		)
		require.NoError(t, c.Init(t.Context()))

		err := c.HealthCheck(t.Context())

		assert.NoError(t, err)
	})

	t.Run("health checks services that implement HealthCheck on the definition", func(t *testing.T) {
		t.Parallel()

		service1 := NewMockLifecycleService(t, "service1")
		service2 := NewMockLifecycleService(t, "service2")
		service3 := NewMockLifecycleService(t, "service3")

		service1.MockIniter.EXPECT().Init(t.Context()).Return(nil)
		service2.MockIniter.EXPECT().Init(t.Context()).Return(nil)
		service3.MockIniter.EXPECT().Init(t.Context()).Return(nil)

		service1.MockHealthChecker.EXPECT().HealthCheck(t.Context()).Return(nil)
		service2.MockHealthChecker.EXPECT().HealthCheck(t.Context()).Return(nil)
		service3.MockHealthChecker.EXPECT().HealthCheck(t.Context()).Return(nil)

		c := pal.NewContainer(&pal.Pal{}, service1, service2, service3)
		require.NoError(t, c.Init(t.Context()))

		err := c.HealthCheck(t.Context())

		assert.NoError(t, err)
	})

	t.Run("returns error when service health check fails", func(t *testing.T) {
		t.Parallel()

		service := NewMockLifecycleService(t, "service1")
		service.MockIniter.EXPECT().Init(t.Context()).Return(nil)
		service.MockHealthChecker.EXPECT().HealthCheck(t.Context()).Return(errTest)

		c := pal.NewContainer(&pal.Pal{}, service)
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

		service1 := NewMockService(t, "service1")
		service2 := NewMockService(t, "service2")

		c := pal.NewContainer(&pal.Pal{}, service1, service2)
		require.NoError(t, c.Init(t.Context()))

		result := c.Services()

		assert.Len(t, result, 2)
		assert.Contains(t, result, "service1")
		assert.Contains(t, result, "service2")
	})

	t.Run("returns empty map for empty container", func(t *testing.T) {
		t.Parallel()

		c := pal.NewContainer(&pal.Pal{})

		result := c.Services()

		assert.Empty(t, result)
	})
}

var _ pal.ServiceDef = (*mockLifecycleService)(nil)
