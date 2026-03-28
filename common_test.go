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

type Pinger interface {
	Ping()
}

type Pinger1 struct{}

func (p *Pinger1) Ping() {}

type Pinger2 struct{}

func (p *Pinger2) Ping() {}

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

// factoryMultiLabel is a catch-all payload for multi-arity factory tests (service_factory0 through 5).
type factoryMultiLabel struct {
	S0, S1, S2 string
	I0, I1     int
	B          bool
	R          rune
}

func newPal(services ...pal.ServiceDef) *pal.Pal {
	return pal.New(services...).
		InitTimeout(time.Second).
		HealthCheckTimeout(time.Second).
		ShutdownTimeout(3 * time.Second)
}
