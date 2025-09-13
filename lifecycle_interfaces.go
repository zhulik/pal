package pal

import "context"

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
	// 1. Whena termination signal is received or the context passed to Pal.Run() is canceled, Pal initiates the shutdown sequence. Services
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
