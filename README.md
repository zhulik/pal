# Pal

[![GoDoc](https://godoc.org/github.com/zhulik/pal?status.svg)](https://pkg.go.dev/github.com/zhulik/pal)
![Build Status](https://github.com/zhulik/pal/actions/workflows/ci.yml/badge.svg)
[![License](https://img.shields.io/github/license/zhulik/pal)](./LICENSE)

Pal is an opinionated [IoC](https://en.wikipedia.org/wiki/Inversion_of_control) framework for Go.

## Motivation

TODO: write me

## Goals
- Nondisruptive API: 
  - You can integrate pal with any app, even if it already uses another IoC frameworks.
  - Even though migration an existing app to pal may require some app redesign, you can do it gradually, 
    migrating one module at a time.
  - Pal tries not to leak into service implementations, so in most of the cases you won't need even struct tags 
    in your services. But if you really need to interact with pal within your services, you can do it.

- Versatility: 
  - You can use pal to build any kind of app: cli, dbms, web, videogames, anything.
  - Pal is aware of other IoC tools and application frameworks, so it tries to coexist with them 
    rather than conflict.

- Safety:
  - When following simple rules, pal never explodes in runtime in the middle of the night. It will only explode
    during initialization, so you can catch it immediately after deployment. Unfortunately, we can't check everything in 
    compile time.
  - Pal tries its best to gracefully shut down the app when interrupted
  - Pal does not try to recover from errors. It's user's responsibility to design resilient services, and it's the 
    execution environment's responsibility to restart the crashed app.
  - Pal is aware of contexts, all service lifetimes callbacks have timeouts: inits, health checks and shutdowns.
    User is forced to configure these timeouts.
  - After initialized, pal is goroutine-safe.

## Non-goals
- Performance: it's assumed that pal is only active during app initialization and shutdown, all other time it only 
  performs periodic health checks. Thus pal's initialization and shutdown should not be blazing fast, it should be *fast enough*. 
  Using factory services is more expensive than using singleton services, but should generally be *fast enough*.

- Extensibility and configurability: pal is not designed to be the most flexible IoC framework, only *flexible enough*
  to reach its goals.

- Lightweightness: while looking simple and having minimalistic API, pal is not that simple inside. It uses **reflection**
  and some other dirty tricks so you don't have to.

- Fool-proofness: even though pal performs some configuration validation, it does not protect from making all possible
  mistakes. For instance, it's user's responsibility to make sure each registered service uses a unique interface.
