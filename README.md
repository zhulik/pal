# Pal

[![GoDoc](https://godoc.org/github.com/zhulik/pal?status.svg)](https://pkg.go.dev/github.com/zhulik/pal)
![Build Status](https://github.com/zhulik/pal/actions/workflows/ci.yml/badge.svg)
[![License](https://img.shields.io/github/license/zhulik/pal)](./LICENSE)

Pal is an opinionated [IoC](https://en.wikipedia.org/wiki/Inversion_of_control) framework for Go.

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
  - Pal should provide tools to simplify testing as much as possible.[TODO]

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

  Each service can implement any of optional service interfaces:
  - [Initer](./interfaces.go#L31) — if the service requires initialization such as connecting to a database.
  - [Shutdowner](./interfaces.go#L20) — if the service requires finalization such as closing connection to a database.
  - [HealthChecker](./interfaces.go#L10) — if the service can be inspected by checking its database connection status.

## Examples:

Examples can be found here:
- [example_container_test.go](./example_container_test.go)
- [example_pal_test.go](./example_pal_test.go)
 