# Pal

[![GoDoc](https://godoc.org/github.com/zhulik/pal?status.svg)](https://pkg.go.dev/github.com/zhulik/pal)
![Build Status](https://github.com/zhulik/pal/actions/workflows/ci.yml/badge.svg)
[![License](https://img.shields.io/github/license/zhulik/pal)](./LICENSE)

Pal is an opinionated [IoC](https://en.wikipedia.org/wiki/Inversion_of_control) framework for Go.

Pal is almost feature complete and is already used in production. However, contributions with examples, documentation updates and
new features are very welcome.

## Motivation

The existing IoC frameworks are either too "heavy" and too flexible for my purpose like [fx](https://github.com/go-uber/fx)
or [wire](https://github.com/google/wire) or too low level and restrictive like [do](https://github.com/samber/do). All
of them share the same trait: you need to design your app with them in mind.

For the past few years I've been using [do](https://github.com/samber/do) in my personal and commercial projects, and it
worked quite well. However, it does not have reflection-based injection, and its API does not really help with
implementing your own. This leads to huge amounts of boilerplate code where you fetch your dependencies from the injector
over and over again.

Pal inherits some design decisions from `fx` and `do` and tries to offer the most nondisruptive IoC experience if you follow
a few rules described below.

### Goals

- Nondisruptive API:
  - You can integrate pal with any app, even if it already uses another IoC framework.
  - Even though migrating an existing app to pal may require some app redesign, you can do it gradually,
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
  - Pal tries its best to gracefully shut down the app when interrupted.
  - Pal does not try to recover from errors. It's user's responsibility to design resilient services, and it's the
    execution environment's responsibility to restart the crashed app.
  - Pal is aware of contexts, all service lifetime callbacks have timeouts: inits, health checks and shutdowns.
    Users are forced to configure these timeouts.
  - After initialization, pal is goroutine-safe.

### Non-goals

- Performance: it's assumed that pal is only active during app initialization and shutdown, all other time it only
  performs periodic health checks. Thus, pal's initialization and shutdown should not be blazing fast, it should be *fast enough*.
  Using factory services is more expensive than using singleton services, but should generally be *fast enough*.

- Extensibility and configurability: pal is not designed to be the most flexible IoC framework, only *flexible enough*
  to reach its goals.

- Lightweightness: while looking simple and having a minimalistic API, pal is not that simple inside. It uses **reflection**
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
    - [Runner](./lifecycle_interfaces.go#L55) — is a special singleton service that runs in the background. Pal runs such services in
      the background. Any app should have at least one Runner. For a web service it would be the place where you run
      `http.ListenAndServe(...)`. For a CLI - where the main logic is. `Pal.Run()` exits when all runners are finished.
      Run() will exit immediately if no runners are registered.
      If another framework already uses `Run` on your type, implement [PalRunner](./lifecycle_interfaces.go#L93) and `PalRun` instead (same behavior).
  - Factory service — a special type of service. Unlike singletons, a new instance of a factory service is created every
    time it is invoked.
  - Const service — a service that wraps an existing instance. It's useful for registering already created objects as services.

  Each service can implement optional lifecycle interfaces. Prefer the standard names when they do not conflict with other frameworks; otherwise use the Pal-prefixed pair (same semantics, different method names):
  - [Initer](./lifecycle_interfaces.go#L39) / [PalIniter](./lifecycle_interfaces.go#L84) — one-time setup after dependencies are injected (e.g. open a DB pool). Use `PalInit` when `Init` is already taken.
  - [Shutdowner](./lifecycle_interfaces.go#L20) / [PalShutdowner](./lifecycle_interfaces.go#L75) — cleanup (e.g. close connections). Use `PalShutdown` when `Shutdown` is already taken.
  - [HealthChecker](./lifecycle_interfaces.go#L5) / [PalHealthChecker](./lifecycle_interfaces.go#L66) — liveness logic for probes. Use `PalHealthCheck` when `HealthCheck` is already taken.

  **Dispatch order** when more than one mechanism could apply: lifecycle hooks (`ToInit`, `ToShutdown`, `ToHealthCheck`) run first; then Pal-prefixed methods if implemented; then the standard `Init` / `Shutdown` / `HealthCheck` methods. For background work, `PalRun` is preferred over `Run` when both exist. In normal use, implement **one** style per phase (standard **or** Pal-prefixed), not both on purpose.

## API Functions

Pal provides several functions for registering services:

- `Provide[T any](value T)` - Registers an instance of service.
- `ProvideFn[T any](fn func(ctx context.Context) (T, error))` - Registers a singleton service created using the provided function.
- `ProvideFactory{0-5}[I any, T any, {0-5}P any](fn func(ctx context.Context, {0-5}P args) (T, error)))` - Registers a factory service created using the provided function with given amount of arguments.
- `ProvideList(...ServiceDef)` - Registers multiple services at once, useful when splitting apps into modules, see [example](./examples/web)
- There are also `Named` versions of `Provide` functions, they can be used along with `name` tag and `Named` versions `Invoke` functions if you want to give your services explicit names.

Pal also provides functions for retrieving services:

- `Invoke[T](ctx, invoker, args...)` - Retrieves or creates an instance of type `T` from the container, factory services may require arguments.
- `InvokeAs[T, C](ctx, invoker, args...)` - A wrapper around `Invoke`, casts the invoked service to `C`, and returns an error if casting fails.
- `InvokeByInterface[I](ctx, invoker, args...)` - Retrieves the only service that implements the given interface `I`.
  Returns an error if there are zero or more than one service implementing the interface or if `I` is not an interface.
  **Note:** do not overuse this function as it gets slower the more services you have.
- `Build[S](ctx, invoker)` - Creates an instance of S, resolves its dependencies, injects them into its fields.
- `InjectInto[S](ctx, invoker, *S)` - Resolves S's dependencies and injects them into its fields.
- There are `Named` versions of `Invoke` functions that allow retrieving services by their explicit names.

All these functions accept nil as invoker, in this case, a Pal instance will be extracted from the context.
Pal automatically adds itself into contexts passed to `Init`, `Shutdown`, and `Run` under the `pal.CtxValue` key.
You can extract it manually with `pal.FromContext`

## Service Types

Pal supports several types of services, each designed for different use cases:

### Singleton Services

Singleton services are created once during application initialization and reused throughout the application's lifetime. They are ideal for:

- Database connections and clients
- HTTP clients and servers
- Configuration objects
- Business logic services
- Any stateful component that should be shared

**Registration:**

```go
// Register a singleton service
pal.Provide[MyService](&MyServiceImpl{})

// Register a singleton service using a factory function
pal.ProvideFn[MyService](func(ctx context.Context) (MyServiceImpl, error) {
    return &MyServiceImpl{}, nil
})
```

### Factory Services

Factory services create a new instance every time they are invoked. They may accept up to 5 arguments. Factories that accept
arguments cannot be explicit dependencies of other services. They are perfect for:

- Stateless components
- Request-scoped objects
- Objects that need different configurations per use
- Components that should not be shared

**Registration:**

```go
// Register a factory service with no arguments
pal.ProvideFactory0[MyService](func(ctx context.Context) (*MyServiceImpl, error) {
    return &MyServiceImpl{}, nil
})

// Register a factory service with arguments
pal.ProvideFactory2[MyService](func(ctx context.Context, url string, timeout time.Duration) (*MyServiceImpl, error) {
    return &MyServiceImpl{URL: url, Timeout: timeout}, nil
})
```

#### Invocation

There are 2 ways to invoke a factory service:

- manual invocation:

  ```go
  pal.Invoke[MyService](ctx, p, "https://example.com", timeout)
  ```

  this way **must never** be used during initialization as Pal does not know that your service depends on a factory service and the factory service
  may not be yet initialized.

- invocation using injected factory function:

  ```go
    type SomeService struct {
      ...
      // parameters of a factory function must match the parameters of the function passed to pal.ProvideFactory
      // but the return value must match the first type argument pal.ProvideFactory
      CreateMyService(ctx context.Context, url string, timeout time.Duration) (MyService, error) 
      ...
    }
  ```

  This way Pal can see that `SomeService` depends on `MyService` and adjust the initialization process accordingly.
  It is safe to call `CreateMyService` from `MyService.Init()`.

### Const Services

Const services wrap existing instances. They are useful for:

- Registering third-party objects
- Wrapping existing instances that don't implement pal interfaces
- Testing with mock objects

**Registration:**

```go
// Register a const service
existingInstance := &MyServiceImpl{}
pal.ProvideConst[MyService](existingInstance)
```

### Runner Services

Runner services are special singleton services that run in the background. They implement the `Runner` interface and are used for:

- HTTP servers
- Message consumers
- Background workers
- Long-running processes

**Example:**

```go
type HTTPServer struct {
    // dependencies...
}

func (s *HTTPServer) Run(ctx context.Context) error {
    // Start HTTP server and run until context is canceled
    return http.ListenAndServe(":8080", s.router)
}
```

## Tags

Pal supports 3 struct tags:

- `pal:"skip"` - fields marked with this tag won't be injected.
- `pal:"match_interface"` - `InvokeByInterface` will be used to inject this dependency
- `pal:"name=<name>"` - a service will be invoked by its explicit name.

## Lifecycle Hooks

Pal provides lifecycle hooks that allow you to customize service behavior without implementing the full lifecycle interfaces. Hooks take precedence over interface methods and are useful for:

- Adding custom initialization logic
- Implementing custom shutdown procedures
- Adding health check functionality
- Wrapping existing objects with lifecycle behavior

### Available Hooks

- **ToInit** - Called during service initialization, after dependencies are injected
- **ToShutdown** - Called during service shutdown, before dependencies are shut down
- **ToHealthCheck** - Called during health checks

### Using Hooks

Even though pal supports hooks, using lifecycle management methods is recommended in the first place as they keep
lifecycle management code closer to the resources it manages.

Hooks can be used with any service type and provide a flexible way to add lifecycle behavior:

```go
// With const services
pal.ProvideConst[MyService](existingInstance).
    ToInit(func(ctx context.Context, service MyService, pal *pal.Pal) error {
        // Custom initialization logic
        return service.Connect()
    }).
    ToShutdown(func(ctx context.Context, service MyService, pal *pal.Pal) error {
        // Custom shutdown logic
        return service.Disconnect()
    }).
    ToHealthCheck(func(ctx context.Context, service MyService, pal *pal.Pal) error {
        // Custom health check logic
        return service.Ping()
    })

// With function-based services
pal.ProvideFn[MyService](func(ctx context.Context) (*MyServiceImpl, error) {
    return &MyServiceImpl{}, nil
}).
    ToInit(func(ctx context.Context, service MyService, pal *pal.Pal) error {
        return service.Initialize()
    })
```

## Examples

Examples can be found here:

- [example_container_test.go](./example_container_test.go)
- [example_pal_test.go](./example_pal_test.go)

## Example apps

- [Web Server](./examples/web) - Demonstrates how to build a web server using Pal.
- [CLI Application](./examples/cli) - Illustrates how to structure a command-line application using Pal.
- [Dependency management using hooks](./examples/hooks) - Illustrates how to use hooks.
- [Factories](./examples/factories) - Illustrates how to use factories.

## Service and container lifecycle

The lifecycle of services and the container in Pal follows a well-defined sequence:

1. **Registration**: Services are registered with Pal using the `Provide*` functions.

2. **Initialization**:
   - When `Pal.Init()` is called, Pal builds a dependency graph of all registered services.
   - Services are initialized in dependency order (dependencies first).
   - For each service, Pal:
     - Creates an instance
     - Injects dependencies into its fields
     - Calls `ToInit` hook if specified; otherwise `PalInit()` if it implements [PalIniter](./lifecycle_interfaces.go#L84), otherwise `Init()` if it implements [Initer](./lifecycle_interfaces.go#L39)
     - If `ToInit` is specified, neither `PalInit` nor `Init` is called.

3. **Running**:
   - After initialization, Pal starts all services that implement [Runner](./lifecycle_interfaces.go#L55) or [PalRunner](./lifecycle_interfaces.go#L93) in background goroutines.
   - Pal calls `PalRun` or `Run` (respecting the same precedence as above when both exist) with a context that will be canceled during shutdown.

4. **Health Checking**:
   - Developers can use `Pal.HealthCheck()` to initiate the health check sequence. In a web application it should be called
     from the `/health` handler which can be used as a liveness probe.
   - Services may implement health checks via `ToHealthCheck`, or via `PalHealthCheck` ([PalHealthChecker](./lifecycle_interfaces.go#L66)), or via `HealthCheck` ([HealthChecker](./lifecycle_interfaces.go#L5)).
   - If `ToHealthCheck` is specified, neither `PalHealthCheck` nor `HealthCheck` is called.
   - If any service returns an error, Pal initiates a graceful shutdown.
5. **Shutdown**:
   - When `Pal.Shutdown()` is called or a termination signal is received, Pal initiates the shutdown sequence.
   - Pal cancels the context for all running services (Runners) and awaits for runners to finish.
   - Pal shuts down dependencies in reverse to initialization order. If `ToShutdown` is set, it runs; otherwise `PalShutdown` or `Shutdown` is used in that precedence order.
   - If `ToShutdown` is specified, neither `PalShutdown` nor `Shutdown` is called.
   - If all services shut down successfully, `Pal.Run()` returns nil, otherwise it returns the collected errors.

## Additional features

### Integration with slog

Pal can automatically inject `*slog.Logger` to your services. To enable this behavior call `InjectSlog()`. Pal
will automatically add the name of service to the `component` attribute of the logger. Attributes added to the
injected logger can be customized by passing arguments to `InjectSlog()`.

### Embedded healthcheck server

Pal includes an embedded healthcheck server so you don't have to implement it yourself. Just call
`RunHealthCheckServer(":8081", "/healthz")` and specify addr and path and the server will start on the specified
addr and respond on GET requests on the specified path. It never responds with a body and does not add any headers
beyond the default ones. It may return one of 4 status codes:

- **200** - all services are healthy
- **404** - wrong path requested
- **405** - wrong HTTP method is used
- **500** - one or more services are unhealthy

### Service dependency inspection

Pal includes a built-in inspection module that provides a web interface to visualize your service dependency graph. This is useful for understanding the structure of your application and debugging dependency issues.

To enable the inspection server, register the inspect service:

```go
import "github.com/zhulik/pal/inspect"

// Register with default port (24242)
pal.ProvideList(inspect.Provide())

// Or specify a custom port
pal.ProvideList(inspect.Provide(8080))
```

The inspection server provides two endpoints:

- **`/pal/tree`** - Interactive HTML visualization of the dependency graph
- **`/pal/tree.json`** - JSON representation of the dependency graph for programmatic access

The visualization shows:

- Service nodes with their types (singleton, factory)
- Dependency relationships between services
- Service capabilities (standard and Pal-prefixed lifecycle interfaces: init, run, health check, shutdown)

## Best practices

To get the most out of Pal, follow these best practices:

1. **Service Design**:
   - Design services as small, focused components with a single responsibility.
   - Use interfaces to define service contracts, especially for services that might have multiple implementations.
   - Implement the optional lifecycle interfaces ([Initer](./lifecycle_interfaces.go#L39), [Shutdowner](./lifecycle_interfaces.go#L20), [HealthChecker](./lifecycle_interfaces.go#L5), or the [Pal-prefixed](./lifecycle_interfaces.go#L66) alternatives when names clash) when appropriate.
   - Use `ProvideConst` and `ProvideFn*` functions with `ToShutdown` hook to register services without dedicated
     interfaces and struct wrappers.

2. **Dependency Management**:
   - Use singleton services for stateful components like database connections, HTTP clients, etc.
   - Use factory services for stateless components that need to be created on demand.
   - Prefer constructor injection (via fields) over service locator pattern (directly using `Pal.Invoke()`).

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
   - Use ToInit hooks to configure services that don't have their own configuration mechanism.

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

## Development

`asdf` is the recommended package manager for development tools. Tool versions are pinned in `.tool-versions`.

Basic setup:

```bash
asdf plugin add golang
asdf plugin add golangci-lint
asdf plugin add task
asdf install
task
```

If `task` runs successfully, your local development setup is ready.

## Contributing

Contributions are welcome.

1. Fork it, implement your feature and open a PR.
2. Wait for review.
3. Address possible comments.
4. ???
5. You're a contributor now, many thanks!
