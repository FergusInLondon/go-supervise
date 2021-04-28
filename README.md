# Go Supervise [![Go Report Card](https://goreportcard.com/badge/github.com/FergusInLondon/go-supervise)](https://goreportcard.com/report/github.com/FergusInLondon/go-supervise) [![Github Workflow](https://github.com/FergusInLondon/go-supervise/actions/workflows/go.yml/badge.svg)](https://github.com/FergusInLondon/go-supervise/actions) [![Go Reference](https://pkg.go.dev/badge/go.fergus.london/go-supervise.svg)](https://pkg.go.dev/go.fergus.london/go-supervise)

A simple implementation of Erlang/OTP's *Supervisor* pattern for Go. A Supervisor is a simple mechanism for improving fault-tolerance of concurrent processes; minimising interruptions in the event of any errors.

![Example](https://d33wubrfki0l68.cloudfront.net/00f3a22cb236e9b62c6944440d74d2df16e9f277/92cc1/diagrams/go-pipeline.png)

## Why?

I love the idea of Erlang/OTP's concurrency model, and specifically Supervision Trees. ¯\\_(ツ)_/¯

It's also a pattern I've found myself implementing in the past, and one that I've recently written about. See [the original blog post](https://fergus.london/blog/lessons-in-concurrency-from-erlang/).

## Usage

Please see the automatically generated [go documentation](https://pkg.go.dev/go.fergus.london/go-supervise) in addition to the [examples directory](./examples).

### NOTE

- Workers - or `Supervisables` - **must** ensure that they capture panics via `recover()` and that they close the provided channel before closing. This can be done in one single deferred function. See the examples for more information.

## Development

See the `Makefile` for tasks; but there's tests, linting, and docs.

### To Do...?

- If I get around to it I'll actually make it tree like; at the moment there's the concept of a `Supervisor` which is likely enough for most pipeline. However nesting Supervisors to make a multi-layer tree will requiring wrapping the individual structs in a function that adheres to `Supervisable`.

- Notifications that a Supervisor has completed. This will most likely be done via a `sync.WaitGroup`; and perhaps some hacking around the user provided `sync.WaitGroup`.

## License

This repository is distributed under the terms of the MIT license; if - *for some unknown reason* - you wish to use this, then you're free to do as you like as per the terms of that license, with the caveat that the license must be distributed in conjunction with this codebase. Additionally - no warranty - if this messes yo' shit up then do not blame me.

See [./LICENSE](./LICENSE).
