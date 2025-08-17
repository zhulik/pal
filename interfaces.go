package pal

import (
	"context"
)

// RunConfiger is an optional interface that can be implemented by a runner to tell Pal how to handle it.
type RunConfiger interface {
	RunConfig() *RunConfig
}

// ServiceDef is a definition of a service. In the case of a singleton service, it also holds the instance.
// This interface combines all the lifecycle interfaces (Initer, HealthChecker, Shutdowner, Runner)
// and adds methods specific to service definition and management.
type ServiceDef interface {
	Initer
	HealthChecker
	Shutdowner
	Runner
	RunConfiger

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

	// InjectInto injects services into the fields of the target struct.
	// It looks at each field's type and tries to find a matching service in the container.
	// Only exported fields can be injected into.
	InjectInto(ctx context.Context, target any) error
}
