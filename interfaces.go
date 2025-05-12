package pal

import (
	"context"
)

type LifecycleHook[T any] func(ctx context.Context, service *T) error

// HealthChecker is an optional interface that can be implemented by a service.
type HealthChecker interface {
	// HealthCheck is being called when pal is checking the health of the service.
	// If returns an error, pal will consider the service unhealthy and try to gracefully Shutdown the app,
	// Pal.Run() will return an error.
	// ctx has a timeout and only being canceled if it is exceeded.
	// DOC: document healthcheck process
	HealthCheck(ctx context.Context) error
}

// Shutdowner is an optional interface that can be implemented by a service.
type Shutdowner interface {
	// Shutdown is being called when pal is shutting down the service.
	// If returns an error, pal will consider this service unhealthy, but will continue to Shutdown the app,
	// Pal.Run() will return an error.
	// ctx has a timeout and only being canceled if it is exceeded.
	// If all of the services successfully Shutdown, Pal.Run will return nil.
	// DOC: document Shutdown process
	Shutdown(ctx context.Context) error
}

// Initer is an optional interface that can be implemented by a service.
type Initer interface {
	// Init is being called when pal is initializing the service, after all the dependencies are injected.
	// If returns an error, pal will consider the service unhealthy and try to gracefully Shutdown already initialized services.
	// DOC: document Init process
	Init(ctx context.Context) error
}

// Runner is a service that can be startRunners in a background goroutine.
// If a service implements it, pal will startRunners thin this method is a background goroutine when app is initialized.
// Can be a one-off or long-running. Services implementing this interface are initialized eagerly.
type Runner interface {
	// Run is being called in a background goroutine when pal is initializing the service, after Init() is called.
	// ctx never being canceled and should be used as the root context for the background job.
	Run(ctx context.Context) error
}

// ServiceDef is a definition of a service. In the case of a singleton service, it also holds the instance.
type ServiceDef interface {
	Initer
	HealthChecker
	Shutdowner
	Runner

	// Name returns a name of the service, this will be used to identify the service in the container.
	Name() string

	// Make only creates a new instance of the service, it doesn't initialize it. Used only to build the dependency DAG.
	Make() any

	// Instance returns a stored instance in the case of singleton service and a new instance in the case of factory.
	Instance(ctx context.Context) (any, error)

	// Validate validates the definition.
	// TODO: make it optional and implement for shipped services
	Validate(_ context.Context) error
}

type Invoker interface {
	Invoke(ctx context.Context, name string) (any, error)
	InjectInto(ctx context.Context, target any) error
}
