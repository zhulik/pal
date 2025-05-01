package pal

import (
	"context"
)

// HealthChecker is an optional interface that can be implemented by a service.
type HealthChecker interface {
	// HealthCheck is being called when pal is checking the health of the service.
	// If returns an error, pal will consider the service unhealthy and try to gracefully shutdown the app,
	// Pal.Run() will return an error.
	// ctx has a timeout and only being canceled if it is exceeded.
	// DOC: document healthcheck process
	HealthCheck(ctx context.Context) error
}

// Shutdowner is an optional interface that can be implemented by a service.
type Shutdowner interface {
	// Shutdown is being called when pal is shutting down the service.
	// If returns an error, pal will consider this service unhealthy, but will continue to shutdown the app,
	// Pal.Run() will return an error.
	// ctx has a timeout and only being canceled if it is exceeded.
	// If all of the services successfully shutdown, Pal.Run will return nil.
	// DOC: document shutdown process
	Shutdown(ctx context.Context) error
}

// Initer is an optional interface that can be implemented by a service.
type Initer interface {
	// Init is being called when pal is initializing the service, after all the dependencies are injected.
	// If returns an error, pal will consider the service unhealthy and try to gracefully shutdown already initialized services.
	// DOC: document init process
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

// Service is a factory for creating a service.
type Service interface {
	// Make only creates a new instance of the service, it doesn't initialize it. Used only to build the dependency DAG.
	Make() any

	// Initialize - when called, singleton services should create its instances.
	Initialize(ctx context.Context) error

	// Instance returns a stored instance in the case of singleton service and a new instance in the case of factory.
	Instance(ctx context.Context) (any, error)

	// Name returns a name of the service, this will be used to identify the service in the container.
	Name() string

	// IsSingleton returns true if the service is a singleton and should be cached and reused.
	IsSingleton() bool

	// IsRunner returns true if the service is a runner.
	IsRunner() bool

	// Validate validates the service factory.
	Validate(_ context.Context) error
}

type loggerFn func(fmt string, args ...any)
