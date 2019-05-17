[![This project is considered experimental](https://img.shields.io/badge/Status-experimental-red.svg)](https://arp242.net/status/experimental)
[![Build Status](https://travis-ci.org/Carpetsmoker/log.svg?branch=master)](https://travis-ci.org/Carpetsmoker/log)
[![GoDoc](https://godoc.org/github.com/Carpetsmoker/log?status.svg)](https://godoc.org/github.com/Carpetsmoker/log)

Martin's logging library.

Canonical package path: `arp242.net/log`.


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
log.Debug("bar")                  // Print debug logs only for module "bar".
log.Module("foo").Debug("w00t")   // Prints nothing.
log.Module("bar").Debug("w00t")   // 15:56:55 w00t
```

Trace logs are kind of like debug logs, but are also printed when there is an
error.

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
