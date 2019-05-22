[![This project is considered experimental](https://img.shields.io/badge/Status-experimental-red.svg)](https://zgo.at/status/experimental)
[![Build Status](https://travis-ci.org/zgoat/log.svg?branch=master)](https://travis-ci.org/zgoat/log)
[![codecov](https://codecov.io/gh/zgoat/log/branch/master/graph/badge.svg)](https://codecov.io/gh/zgoat/log)
[![GoDoc](https://godoc.org/github.com/zgoat/log?status.svg)](https://godoc.org/github.com/zgoat/log)

Go logging library. Canonical import path: `zgo.at/log`.

The main goal is to offer a friendly API which gets out of your way.

Getting the maximum possible amount of performance or zero-allocations are not a
goal, although simple benchmarks show it should be Fast Enough™ for most
purposes (if not, there are a few max-performance libraries already).

Usage
-----

Basics:

```go
log.Print("foo")                  // 15:55:17 foo
log.Printf("foo %d", 1)           // 15:55:17 foo 1
log.Error(err)                    // 15:55:17 oh noes
log.Errorf("foo %d", 1)           // 15:55:17 foo 1
```

This all does what you expect: output a message to stdout or stderr.

You can add module information and fields for extra information:

```go
l := log.Module("test")
l.Print("foo")                    // 15:56:12 test: foo

l = l.Fields(log.F{"foo": "bar"})
l.Print("foo")                    // 15:56:55 test: foo key="val"
```

Debug logs are printed only for modules marked as debug:

```go
log.Module("bar").Debug("w00t")   // Prints nothing (didn't enable module "bar").
log.Debug("bar")                  // Enable debug logs only for module "bar".
log.Module("bar").Debug("w00t")   // 15:56:55 w00t
```

Trace logs are like debug logs, but are also printed when there is an error.

```go
l := log.Module("foo")
l.Trace("useful info")
l.Error(err)
l.ResetTrace()                    // Remove all traces.
```

This is pretty useful for adding context to errors without clobbering your
general log with mostly useless info.

Configuration
-------------

Configuration is done by setting the `Config` variable (usually during
initialisation of your app).

It's not possible to configure individual logger instances. It's not often
needed, and adds some complexity.
