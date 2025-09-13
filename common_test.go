package pal_test

import (
	"errors"
	"testing"
	"time"

	"github.com/zhulik/pal"
)

var (
	errTest  = errors.New("test error")
	errTest2 = errors.New("test error 2")
)

// TestServiceInterface is a simple interface for testing
type TestServiceInterface any

// TestServiceStruct is a test helper struct
type TestServiceStruct struct {
	*MockHealthChecker
	*MockIniter
	*MockShutdowner
}

func NewMockTestServiceStruct(t *testing.T) *TestServiceStruct {
	return &TestServiceStruct{
		MockHealthChecker: NewMockHealthChecker(t),
		MockIniter:        NewMockIniter(t),
		MockShutdowner:    NewMockShutdowner(t),
	}
}

// RunnerServiceStruct is a test helper struct
type RunnerServiceStruct struct {
	*MockRunner
	*MockRunConfiger
}

func NewMockRunnerServiceStruct(t *testing.T) *RunnerServiceStruct {
	return &RunnerServiceStruct{
		MockRunner:      NewMockRunner(t),
		MockRunConfiger: NewMockRunConfiger(t),
	}
}

// DependentStruct is a struct with a dependency on TestServiceInterface
type DependentStruct struct {
	Dependency *TestServiceStruct
}

func newPal(services ...pal.ServiceDef) *pal.Pal {
	return pal.New(services...).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(3 * time.Second)
}
