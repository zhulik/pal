# Pal web app example

A simple web app that pings google.com from an http endpoint handler. It utilizes pal's dependency lifecycle management:
`Init`, `Shutdown` and `Run`, demonstrates automatic logger injection, embedded healthcheck server and dependency
injection using an interface as
a dependency identifier.

It also demonstrates 2 important concepts:

- Managing a services which do not natively support stopping with a context: `http.Server`
- Modularity: it shows how to split the app into multiple modules
