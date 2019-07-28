[![This project is considered experimental](https://img.shields.io/badge/Status-experimental-red.svg)](https://arp242.net/status/experimental)
[![Build Status](https://travis-ci.org/zgoat/zlog.svg?branch=master)](https://travis-ci.org/zgoat/zlog)
[![codecov](https://codecov.io/gh/zgoat/zlog/branch/master/graph/badge.svg)](https://codecov.io/gh/zgoat/zlog)
[![GoDoc](https://godoc.org/zgo.at/zlog?status.svg)](https://godoc.org/zgo.at/zlog)

Go logging library. Canonical import path: `zgo.at/zlog`.

The main goal is to offer a friendly API which gets out of your way.

Getting the maximum possible amount of performance or zero-allocations are not a
goal, although simple benchmarks show it should be Fast Enough™ for most
purposes (if not, there are a few max-performance libraries already).

Usage
-----

Basics:

```go
zlog.Print("foo")                  // 15:55:17 foo
zlog.Printf("foo %d", 1)           // 15:55:17 foo 1
zlog.Error(err)                    // 15:55:17 oh noes
zlog.Errorf("foo %d", 1)           // 15:55:17 foo 1
```

This does what you expect: output a message to stdout or stderr.

You can add module information and fields for extra information:

```go
log := zlog.Module("test")
log.Print("foo")                    // 15:56:12 test: foo

log = l.Fields(zlog.F{"foo": "bar"})
log.Print("foo")                    // 15:56:55 test: foo key="val"
```

Debug logs are printed only for modules marked as debug:

```go
zlog.Module("bar").Debug("w00t")    // Prints nothing (didn't enable module "bar").
log := zlog.Debug("bar")            // Enable debug logs only for module "bar".
log.Module("bar").Debug("w00t")     // 15:56:55 w00t
```

Trace logs are like debug logs, but are also printed when there is an error:

```go
log := zlog.Module("foo")
log.Trace("useful info")
log.Error(err)
log.ResetTrace()                    // Remove all traces.
```

This is pretty useful for adding context to errors without clobbering your
general log with mostly useless info.

You can also easily print timings; this is intended for development and only
printed for modules marked as debug:

    log := zlog.Debug("long-running").Module("long-running")

    time.Sleep(1 * time.Second)
    log = log.Since("sleep one")    //   long-running 11ms  sleep one

    time.Sleep(20*time.Millisecond)
    log.Since("sleep two")          //   long-running 11ms  sleep two

And finally there is the `Recover()` helper functions to recover from panics:

    go func() {
        defer zlog.Recover()        // Recover panics and report with Error().
        panic("oh noes!")
    }

See [GoDoc](https://godoc.org/zgo.at/zlog) for the full reference.


Configuration
-------------

Configuration is done by setting the `Config` variable (usually during
initialisation of your app).

It's not possible to configure individual logger instances. It's not often
needed, and adds some complexity.

- [zgo.at/zlog-sentry](https://github.com/zgoat/zlog_sentry) – send errors to
  [Sentry](https://sentry.io).
