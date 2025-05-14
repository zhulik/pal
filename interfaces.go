package pal

import (
	"context"
)

// LifecycleHook is a function type that can be registered to run at specific points in a service's lifecycle.
// It receives the service instance and a contex, and can return an error to indicate failure.
// These hooks are typically used with BeforeInit methods to customize service initialization.
type LifecycleHook[T any] func(ctx context.Context, service T) error

// HealthChecker is an optional interface that can be implemented by a service.
type HealthChecker interface {
	// HealthCheck is being called when pal is checking the health of the service.
	// If returns an error, pal will consider the service unhealthy and try to gracefully Shutdown the app,
	// Pal.Run() will return an error.
	// ctx has a timeout and only being canceled if it is exceeded.
	//
	// The healthcheck process works as follows:
	// 1. When Pal.HealthCheck() is called Pal initiates the healthcheck sequence. All services are checked concurrently.
	// 2. If any service returns an error, Pal initiates a graceful shutdown
	// 3. Services can use this method to check their internal state or connections to external systems
	// 4. The context provided has a timeout configured via Pal.HealthCheckTimeout()
	HealthCheck(ctx context.Context) error
}

// Shutdowner is an optional interface that can be implemented by a service.
type Shutdowner interface {
	// Shutdown is being called when pal is shutting down the service.
	// If returns an error, pal will consider this service unhealthy, but will continue to Shutdown the app,
	// Pal.Run() will return an error.
	// ctx has a timeout and only being canceled if it is exceeded.
	// If all the services shutdown successfully, Pal.Run will return nil.
	//
	// The shutdown process works as follows:
	// 1. When Pal.Shutdown() is called or a termination signal is received, Pal initiates the shutdown sequence. Services
	// 	  are shutdown in dependency order.
	// 2. Pal cancels the context for all running services (Runners) and awaits for runners to finish.
	// 3. Pal calls Shutdown() on all services that implement this interface in reverse dependency order
	// 4. Services should use this method to clean up resources, close connections, etc.
	// 5. The context provided has a timeout configured via Pal.ShutdownTimeout()
	// 6. If any service returns an error during shutdown, Pal will collect these errors and return them from Run()
	Shutdown(ctx context.Context) error
}

// Initer is an optional interface that can be implemented by a service.
type Initer interface {
	// Init is being called when pal is initializing the service, after all the dependencies are injected.
	// If returns an error, pal will consider the service unhealthy and try to gracefully Shutdown already initialized services.
	//
	// The initialization process works as follows:
	// 1. During Pal.Init() Pal builds a dependency graph of all registered services
	// 2. Pal initializes services in dependency order.
	// 3. For each service, Pal injects dependencies and then calls Init() if the service implements this interface
	// 4. Services should use this method to perform one-time setup operations like connecting to databases
	// 5. The context provided has a timeout configured via Pal.InitTimeout()
	// 6. If any service returns an error during initialization, Pal will stop the initialization process
	//    and attempt to gracefully shut down any already initialized services
	Init(ctx context.Context) error
}

// Runner is a service that can be started in a background goroutine.
// If a service implements this interface, Pal will start this method in a background goroutine when the app is initialized.
// Can be a one-off or long-running task. Services implementing this interface are initialized eagerly.
type Runner interface {
	// Run is being called in a background goroutine when Pal is initializing the service, after Init() is called.
	// The provided context will be canceled when Pal is shut down, so the service should monitor it and exit gracefully.
	// This method should implement the main functionality of background services like HTTP servers, message consumers, etc.
	// If this method returns an error, Pal will initiate a graceful shutdown of the application.
	Run(ctx context.Context) error
}

// ServiceDef is a definition of a service. In the case of a singleton service, it also holds the instance.
// This interface combines all the lifecycle interfaces (Initer, HealthChecker, Shutdowner, Runner)
// and adds methods specific to service definition and management.
type ServiceDef interface {
	Initer
	HealthChecker
	Shutdowner
	Runner

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
	Instance(ctx context.Context) (any, error)

	// Validate validates the service definition.
	// This is called during container initialization to ensure the service is properly configured.
	// It should check that the service implementation satisfies all required interfaces and constraints.
	Validate(_ context.Context) error
}

// Invoker is an interface for retrieving services from a container and injecting them into structs.
// Both Container and Pal implement this interface, allowing services to be retrieved from either.
type Invoker interface {
	// Invoke retrieves a service by name from the container.
	// Returns the service instance or an error if the service is not found or cannot be initialized.
	Invoke(ctx context.Context, name string) (any, error)

	// InjectInto injects services into the fields of the target struct.
	// It looks at each field's type and tries to find a matching service in the container.
	// Only exported fields can be injected into.
	InjectInto(ctx context.Context, target any) error
}
