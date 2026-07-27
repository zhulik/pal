package pal

import (
	"context"
	"reflect"
)

// RunConfiger is an optional interface a runner may implement to tell Pal whether to wait for it.
// ShouldWaitForRunner is true for a main runner (Pal blocks until it finishes) and false for a
// secondary / fire-and-forget runner. Returning only a bool lets third-party types satisfy this
// interface without importing pal (structural typing).
type RunConfiger interface {
	ShouldWaitForRunner() bool
}

// PalRunConfiger is an alternative interface with the same semantics as [RunConfiger], using a Pal-prefixed method name
// so the type can still implement another type's ShouldWaitForRunner without a clash.
// Prefer [RunConfiger] when method names do not conflict.
// If both PalRunConfiger and [RunConfiger] are implemented, Pal uses [PalRunConfiger.PalShouldWaitForRunner] only.
type PalRunConfiger interface { //nolint:revive
	PalShouldWaitForRunner() bool
}

// ServiceDef is a definition of a service. In the case of a singleton service, it also holds the instance.
// This interface embeds the standard lifecycle interfaces ([Initer], [HealthChecker], [Shutdowner], [Runner]);
// concrete services may instead implement the Pal-prefixed alternatives ([PalIniter], [PalHealthChecker], [PalShutdowner], [PalRunner], [PalRunConfiger]) where method names would otherwise clash.
// It adds methods specific to service definition and management.
type ServiceDef interface {
	Initer
	HealthChecker
	Shutdowner
	Runner

	// ShouldWaitForRunner reports runner scheduling for this service definition.
	// nil means the service is not a runner; otherwise true = main runner and false = secondary.
	// Wrappers default to true when the instance implements Runner/PalRunner but not
	// [RunConfiger]/[PalRunConfiger].
	ShouldWaitForRunner() *bool

	// Name returns a name of the service, this will be used to identify the service in the container.
	// The name is typically derived from the interface type the service implements.
	Name() string

	// Make only creates a new instance of the service, it doesn't initialize it.
	// Used only to build the dependency DAG by analyzing the fields of the returned instance.
	// This method should not have side effects as it may be called multiple times.
	Make() any

	// Instance returns a stored instance in the case of singleton service and a new instance in the case of factory.
	// For singletons, this returns the cached instance after initialization.
	// For factories, this creates and returns a new instance each time.
	Instance(ctx context.Context, args ...any) (any, error)

	// Arguments returns the number of arguments the service expects.
	// This is used to validate the number of arguments passed to the service.
	Arguments() int

	// Dependencies allows services to provide their own dependencies.
	Dependencies() []ServiceDef
}

// Invoker is an interface for retrieving services from a container and injecting them into structs.
// Both Container and Pal implement this interface, allowing services to be retrieved from either.
type Invoker interface {
	// Invoke retrieves a service by name from the container.
	// Returns the service instance or an error if the service is not found or cannot be initialized.
	Invoke(ctx context.Context, name string, args ...any) (any, error)

	// InvokeByInterface retrieves a service by interface from the container.
	// Returns the service instance or an error if the service is not found or cannot be initialized.
	InvokeByInterface(ctx context.Context, iface reflect.Type, args ...any) (any, error)

	// InjectInto injects services into the fields of the target struct.
	// It looks at each field's type and tries to find a matching service in the container.
	// Only exported fields can be injected into.
	InjectInto(ctx context.Context, target any) error
}
