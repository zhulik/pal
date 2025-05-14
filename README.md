# Pal

[![GoDoc](https://godoc.org/github.com/zhulik/pal?status.svg)](https://pkg.go.dev/github.com/zhulik/pal)
![Build Status](https://github.com/zhulik/pal/actions/workflows/ci.yml/badge.svg)
[![License](https://img.shields.io/github/license/zhulik/pal)](./LICENSE)

Pal is an opinionated [IoC](https://en.wikipedia.org/wiki/Inversion_of_control) framework for Go.

Pal is at very early stages of development. Technically, it's almost feature complete, but it needs some testing and 
polishing.

## Motivation

The exising IoC frameworks are either too "heavy" and too flexible for my purpose like [fx](https://github.com/go-uber/fx) 
or [wire](https://github.com/google/wire) or too low level and restrictive like [do](https://github.com/samber/do). All
of them share the same trait: you need to design your app with them in mind.

For the past few years I've been using [do](https://github.com/samber/do) in my personal and commercial projects, and it
worked quite well. However, it does not have reflection-based injection, and it's API does not really help with 
implementing your own. This leads to huge amounts of boilerplate code where you fetch your dependencies from the injector
over and over again.

Pal inherits some design decisions from `fx` and `do` and tries to offer the most nondisruptive IoC experience if you follow
a few rules described below.

### Goals
- Nondisruptive API: 
  - You can integrate pal with any app, even if it already uses another IoC frameworks.
  - Even though migration an existing app to pal may require some app redesign, you can do it gradually, 
    migrating one module at a time.
  - Pal tries not to leak into service implementations, so in most of the cases you won't even need struct tags 
    in your services. But if you really need to interact with pal within your services, you can do it.

- Versatility: 
  - You can use pal to build any kind of app: cli, dbms, web, video games, anything.
  - Pal is aware of other IoC tools and application frameworks, so it tries to coexist with them 
    rather than conflict.

- Testability: 
  - Pal provides tools to simplify testing, such as the ability to register mock services using ProvideConst.
  - The container design allows for easy swapping of real implementations with test doubles.
  - Services can be tested in isolation by creating a test container with only the necessary dependencies.

- Safety:
  - When following simple rules, pal never explodes in runtime in the middle of the night. It will only explode
    during initialization, so you can catch it immediately after deployment. Unfortunately, we can't check everything in 
    compile time.
  - Pal tries its best to gracefully shut down the app when interrupted
  - Pal does not try to recover from errors. It's user's responsibility to design resilient services, and it's the 
    execution environment's responsibility to restart the crashed app.
  - Pal is aware of contexts, all service lifetime callbacks have timeouts: inits, health checks and shutdowns.
    User is forced to configure these timeouts.
  - After initialized, pal is goroutine-safe.

### Non-goals
- Performance: it's assumed that pal is only active during app initialization and shutdown, all other time it only 
  performs periodic health checks. Thus, pal's initialization and shutdown should not be blazing fast, it should be *fast enough*. 
  Using factory services is more expensive than using singleton services, but should generally be *fast enough*.

- Extensibility and configurability: pal is not designed to be the most flexible IoC framework, only *flexible enough*
  to reach its goals.

- Lightweightness: while looking simple and having minimalistic API, pal is not that simple inside. It uses **reflection**
  and some other dirty tricks so you don't have to.

- Fool-proofness: even though pal performs some configuration validation, it does not protect from making all possible
  mistakes. For instance, it's the user's responsibility to make sure each registered service uses a unique interface.

## Glossary
- [IoC](https://en.wikipedia.org/wiki/Inversion_of_control) — Inversion of Control is a design principle in software 
  engineering that transfers control of a process or object from one part of the program to another.
- [Dependency Injection](https://en.wikipedia.org/wiki/Dependency_injection) — specific implementation of the Inversion 
  of Control pattern where objects receive their dependencies through constructor arguments, method calls, or property 
  setters rather than creating them themselves.
- Container — a registry of services within the app. It is responsible for managing service lifecycle.
- Service — is an interface that defines a set of methods or operations. Concrete implementations of these services are 
  responsible for providing specific functionalities within the application. Services can perform tasks on their own or
  can be used by other services. Pal recognizes a few types of services:
  - Singleton service — such services are created only once during application initialization. It can be a client for 
    third party service or a piece of business logic. Most of your services should be singletons.  
    Pal recognizes a special kind of Singleton service:
    - [Runner](./interfaces.go#L41) — is a special singleton service that runs the background. Pal run such services in
      the background. Any app should have at least one Runner. For a web service it would be the place where you run
      `http.ListenAndServe(...)`. For a CLI - where the main logic is. Pal.Run() exits when all runners are finished.
      Run() will exit immediately of no runners registered.
  - Factory service — a special type of service. Unlike singletons, a new instance of a factory service is created every
    time it is invoked.
  - Const service — a service that wraps an existing instance. It's useful for registering already created objects as services.

  Each service can implement any of optional service interfaces:
  - [Initer](./interfaces.go#L31) — if a service requires initialization such as connecting to a database.
  - [Shutdowner](./interfaces.go#L20) — if a service requires finalization such as closing connection to a database.
  - [HealthChecker](./interfaces.go#L10) — if a service can be inspected by checking its database connection status.

## API Functions

Pal provides several functions for registering services:

- `Provide[I, S]()` - Registers a singleton service where `I` is the interface type and `S` is the implementation type.
- `ProvideFn[T](fn)` - Registers a singleton service that is created using the provided function.
- `ProvideFactory[I, S]()` - Registers a factory service where `I` is the interface type and `S` is the implementation type.
- `ProvideFnFactory[T](fn)` - Registers a factory service that is created using the provided function.
- `ProvideConst[T](value)` - Registers an existing instance as a service.

Pal also provides functions for retrieving services:

- `Invoke[T](ctx, invoker)` - Retrieves or creates an instance of type `T` from the container.
- `Build[S](ctx, invoker)` - Creates an instance of S, resolves it's dependencies and injects them into its fields.
- `InjectInto[S](ctx, invoker, *S)` - Resolves S's dependencies and injects them into its fields.

## Examples:

Examples can be found here:
- [example_container_test.go](./example_container_test.go)
- [example_pal_test.go](./example_pal_test.go)

## Example apps:
- [Web Server](./examples/web) - Demonstrates how to build a web server using Pal.
- [Message Consumer](./examples/consumer) - Shows how to implement a message consumer that processes messages from a queue.
- [CLI Application](./examples/cli) - Illustrates how to structure a command-line application using Pal.

## Service and container lifecycle

The lifecycle of services and the container in Pal follows a well-defined sequence:

1. **Registration**: Services are registered with Pal using the `Provide*` functions.

2. **Initialization**:
   - When `Pal.Init()` Pal builds a dependency graph of all registered services.
   - Services are initialized in dependency order (dependencies first).
   - For each service, Pal:
     - Creates an instance
     - Injects dependencies into its fields
     - Calls any registered BeforeInit hooks
     - Calls the service's Init() method if it implements the Initer interface

3. **Running**:
   - After initialization, Pal starts all services that implement the Runner interface in background goroutines.
   - The Run() method of these services is called with a context that will be canceled during shutdown.

4. **Health Checking**:
   - Developers can use `Pal.HealthCheck()` to initiate the healthcheck sequence. In a web application it should be called
     from the `/health` handler which can be used as a liveliness probe.
   - If any service returns an error, Pal initiates a graceful shutdown.

5. **Shutdown**:
   - When Pal.Shutdown() is called or a termination signal is received, Pal initiates the shutdown sequence.
   - Pal cancels the context for all running services (Runners) and awaits for runners to finish.
   - Pal calls Shutdown() on all services that implement the Shutdowner interface in reverse dependency order.
   - If all services shut down successfully, Pal.Run() returns nil, otherwise it returns the collected errors.

## Best practices

To get the most out of Pal, follow these best practices:

1. **Service Design**:
   - Design services as small, focused components with a single responsibility.
   - Use interfaces to define service contracts, especially for services that might have multiple implementations.
   - Implement the optional lifecycle interfaces (Initer, Shutdowner, HealthChecker) when appropriate.
   - Use `ProvideConst` and `ProvideFn*` functions with `BeforeShutdown` hook to register services without dedicated
     interfaces and struc wrappers.

2. **Dependency Management**:
   - Use singleton services for stateful components like database connections, HTTP clients, etc.
   - Use factory services for stateless components that need to be created on demand.
   - Prefer constructor injection (via fields) over service locator pattern (directly using Pal.Invoke).

3. **Error Handling**:
   - Handle errors appropriately in service implementations, especially in Init, Shutdown, and HealthCheck methods.
   - Use the context provided to respect timeouts and cancellation signals.
   - Log errors with appropriate context to aid debugging.

4. **Testing**:
   - Create mock implementations of your service interfaces for testing.
   - Use `ProvideConst` to register mock services in your tests.
   - Test each service in isolation before testing them together.

5. **Application Structure**:
   - Organize your code by domain or feature, not by technical concerns.
   - Keep your main function simple - it should just create Pal, register services, and call Run().
   - Use Runners for long-running processes like HTTP servers or message consumers.

6. **Configuration**:
   - Set appropriate timeouts for initialization, health checking, and shutdown.
   - Register signal handlers to ensure graceful shutdown on termination signals.
   - Use BeforeInit hooks to configure services that don't have their own configuration mechanism.

## Troubleshooting

Here are solutions to common issues you might encounter when using Pal:

1. **Service Not Found**:
   - **Symptom**: `ErrServiceNotFound` error when trying to invoke a service.
   - **Possible Causes**:
     - The service wasn't registered with Pal.
     - The service was registered with a different interface type than the one being requested.
   - **Solution**: Check that you've registered the service with the correct interface type and that the registration happens before the service is invoked.

2. **Service Initialization Failed**:
   - **Symptom**: `ErrServiceInitFailed` error during container initialization.
   - **Possible Causes**:
     - The service's Init method returned an error.
     - A dependency of the service couldn't be initialized.
   - **Solution**: Check the error message for details about which service failed and why. Ensure all dependencies are properly registered and initialized.

3. **Service Invalid**:
   - **Symptom**: `ErrServiceInvalid` error when trying to invoke a service.
   - **Possible Causes**:
     - The service implementation doesn't satisfy the interface it was registered with.
     - Type assertion failed during service retrieval.
   - **Solution**: Ensure the service implementation correctly implements all methods of the interface it's registered with.

4. **Circular Dependencies**:
   - **Symptom**: Error during container initialization about a cycle in the dependency graph.
   - **Possible Causes**: Two or more services depend on each other, creating a circular dependency.
   - **Solution**: Refactor your services to break the circular dependency. Consider using a factory service or restructuring your code.

5. **Timeout During Initialization/Shutdown**:
   - **Symptom**: Panic with "initialization timed out" or "shutdown timed out".
   - **Possible Causes**: A service's Init or Shutdown method took longer than the configured timeout.
   - **Solution**: Increase the timeout using `Pal.InitTimeout()` or `Pal.ShutdownTimeout()`, or optimize the service to complete faster.

6. **Context Cancellation Not Respected**:
   - **Symptom**: Services don't shut down gracefully when the context is canceled.
   - **Possible Causes**: The service isn't checking for context cancellation in its Run method.
   - **Solution**: Ensure all long-running operations in your services respect context cancellation by checking `ctx.Done()` regularly.
