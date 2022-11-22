Go logging library. Canonical import path: `zgo.at/zlog`. You will need Go 1.11
or newer. API docs: https://godocs.io/zgo.at/zlog

The main goal is to offer a friendly and ergonomic API. Getting the maximum
possible amount of performance or zero-allocations are not goals, although
simple benchmarks show it should be more than *Fast Enough™* for most purposes
(if not, there are a few max-performance libraries already).

Usage
-----

Basics:

```go
zlog.Print("foo")                  // 15:55:17 INFO: foo
zlog.Printf("foo %d", 1)           // 15:55:17 INFO: foo 1
zlog.Error(errors.New("oh noes"))  // 15:55:17 ERROR: oh noes
zlog.Errorf("foo %d", 1)           // 15:55:17 ERROR: foo 1
```

This does what you expect: output a message to stdout or stderr.

You can add module information and fields for extra information:

```go
log := zlog.Module("test")
log.Print("foo")                    // 15:56:12 test: INFO: foo

log = log.Fields(zlog.F{"foo": "bar"})
log.Print("foo")                    // 15:56:55 test: INFO: foo {foo="bar"}
```

Debug logs are printed only for modules marked as debug:

```go
zlog.Module("bar").Debug("w00t")    // Prints nothing (didn't enable module "bar").
log := zlog.SetDebug("bar")         // Enable debug logs only for module "bar".
log.Module("bar").Debug("w00t")     // 15:56:55 bar: DEBUG: w00t
```

Trace logs are like debug logs, but are also printed when there is an error,
even when debug is disabled for the module:

```go
log := zlog.Module("foo")
log = log.Trace("useful info")
log.Error(errors.New("oh noes"))   // 19:44:26 foo: TRACE: useful info
                                   // 19:44:26 foo: ERROR: oh noes
log = log.ResetTrace()             // Remove all traces.
```

This is pretty useful for adding context to errors without clobbering your
general log with mostly useless info.

You can also easily record timings; this is printed for modules marked as debug:

```go
log := zlog.SetDebug("zzz").Module("zzz")

time.Sleep(1 * time.Second)
log = log.Since("one")             // zzz  1000ms  one

time.Sleep(20*time.Millisecond)
log.Since("two")                   // zzz    20ms  two

// Add timing as fields, always works regardless of Debug.
log.FieldsSince().Print("done")    // 19:48:15 zzz: INFO: done {one="1000ms" two="20ms"}
```

Many functions return a `Log` object. It's important to remember that Log
objects are never modified in-place, so using `log.Trace(..)` without assigning
it is does nothing. This also applies to `SetDebug()`, `Module()`, `Since()`,
etc.

The `Recover()` helper function makes it easier to recover from panics in
goroutines:

```go
go func() {
    defer zlog.Recover()           // Recover panics and report with Error().
    panic("oh noes!")
}()
```

See https://godocs.io/zgo.at/zlog for the full reference.


### Configuration

Configuration is done by setting the `zlog.Config` variable usually during
initialisation of your app.

It's not possible to configure individual logger instances, as it's rarely
needed (but I might change my mind if someone presents a good use-case).

See LogConfig godoc for docs.
