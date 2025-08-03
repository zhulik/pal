# Pal web app example

A simple web app that pings google.com from an http endpoint handler. It utilizes pal's dependency lifecycle managemant:
`Init`, `Shutdown` and `Run`, demostrates automatic logger injection, embedded healthcheck server and dependecy
injection using an interface as
a dependecy identifier.

It also demostrates 2 important concepts:

- Managing a services which do not natively support stopping with a context: `http.Server`
- Modularity: it shows how to split the app into multiple modules
