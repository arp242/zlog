// Package zlog is a logging library.
package zlog // import "zgo.at/zlog"

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"time"
)

// LogConfig is the configuration struct.
type LogConfig struct {
	// Outputs for a Log entry.
	//
	// The default is to print to stderr for errors, and stdout for everything
	// else. Generally you want to keep this as a backup and add additional
	// outputs, instead of replacing this. For example:
	//
	//    zlog.Config.Outputs = append(zlog.Config.Outputs, func(l Log) {
	//        if l.Level != LevelErr { // Only process errors.
	//            return
	//        }
	//
	//        // .. send to external error notification service ..
	//    })
	//
	//    zlog.Config.Outputs = append(zlog.Config.Outputs, func(l Log) {
	//        if l.Level == LevelErr { // Only process non-errors.
	//            return
	//        }
	//
	//        // .. send to external logging service ..
	//    })
	Outputs []OutputFunc

	// Always print debug information for these modules. Debug will be enabled
	// for all modules with the special word "all".
	Debug []string

	// Format function used by the default stdout/stderr output. This takes a
	// Log entry and formats it for output.
	//
	// TODO: the boundary between "outputs" and "zlog internals" are kinda leaky
	// here; it's used for Trace() logs now. Should think about refactoring
	// re-doing this in another way.
	//
	// Maybe add type OutputConfig{ .. } for this (and FmtTime)?
	Format func(Log) string

	// Time/date format as accepted by time.Format(); used in the default
	// Format() function.
	//
	// The default is to just print the time, which works well in development.
	// For production you probably want to use time.RFC3339 or some such.
	//
	// This is used in the standard format() function, not not elsewhere.
	FmtTime string
}

// SetDebug sets the Debug field from a comma-separated list of module names.
func (c *LogConfig) SetDebug(d string) {
	d = strings.TrimSpace(d)
	if d == "" {
		c.Debug = nil
	}
	c.Debug = strings.Split(d, ",")
}

func (c LogConfig) RunOutputs(l Log) {
	for _, o := range c.Outputs {
		o(l)
	}
}

// OutputFunc is an output function, used in Config.Outputs.
type OutputFunc func(Log)

// Config for this package.
var Config LogConfig

func init() {
	Config = LogConfig{
		FmtTime: "15:04:05 ",
		Format:  format,
		Outputs: []OutputFunc{output},
	}
}

// Log levels.
const (
	LevelInfo  = 0
	LevelErr   = 1
	LevelDbg   = 2
	LevelTrace = 3
)

var now = time.Now

type (
	// Log module.
	Log struct {
		Ctx          context.Context
		Msg          string   // Log message; set with Print(), Debug(), etc.
		Err          error    // Original error, set with Error().
		Level        int      // 0: print, 1: err, 2: debug, 3: trace
		Modules      []string // Modules added to the logger.
		Data         F        // Fields added to the logger.
		DebugModules []string // List of modules to debug.
		Traces       []string // Traces added to the logger.

		since    time.Time
		sinceLog F
	}

	// F are log fields.
	F map[string]interface{}
)

// Module adds a module to this Log entry.
//
// You can add multiple Modules.
func Module(m string) Log { return Log{Modules: []string{m}, since: time.Now()} }

func SetDebug(m ...string) Log          { return Log{DebugModules: m} }
func Field(k string, v interface{}) Log { return Log{Data: F{k: v}} }
func Fields(f F) Log                    { return Log{Data: f} }
func Print(v ...interface{})            { Log{}.Print(v...) }
func Printf(f string, v ...interface{}) { Log{}.Printf(f, v...) }
func Error(err error)                   { Log{}.Error(err) }
func Errorf(f string, v ...interface{}) { Log{}.Errorf(f, v...) }

// FieldsRequest adds information from a HTTP request as fields.
func FieldsRequest(r *http.Request) Log { return Log{}.FieldsRequest(r) }

// FieldsLocation records the caller location.
func FieldsLocation() Log { return Log{}.FieldsLocation() }

// ResetTrace removes all trace logs added with Trace() and Tracef().
func (l Log) ResetTrace() Log {
	l.Traces = nil
	return l
}

// Context adds a context to the Log entry.
//
// This isn't used by zlog, and mostly so that outputs can use it if needed.
func (l Log) Context(ctx context.Context) { l.Ctx = ctx }

func (l Log) SetDebug(m ...string) Log {
	l.DebugModules = append(l.DebugModules, m...)
	return l
}

func (l Log) Module(m string) Log {
	l.Modules = append(l.Modules, m)
	l.since = time.Now()
	return l
}

// Fields append data to the Log object.
func (l Log) Fields(f F) Log {
	if l.Data == nil {
		l.Data = f
		return l
	}

	for k, v := range f {
		l.Data[k] = v
	}
	return l
}

// Field sets one data field.
func (l Log) Field(k string, v interface{}) Log {
	return l.Fields(F{k: v})
}

// Print an informational error.
func (l Log) Print(v ...interface{}) {
	l.Msg = fmt.Sprint(v...)
	l.Level = LevelInfo
	Config.RunOutputs(l)
}

// Printf an informational error.
func (l Log) Printf(f string, v ...interface{}) {
	l.Msg = fmt.Sprintf(f, v...)
	l.Level = LevelInfo
	Config.RunOutputs(l)
}

// Error prints an error.
func (l Log) Error(err error) {
	l.Err = err
	l.Level = LevelErr
	Config.RunOutputs(l)
}

// Errorf prints an error.
func (l Log) Errorf(f string, v ...interface{}) {
	l.Err = fmt.Errorf(f, v...)
	l.Level = LevelErr
	Config.RunOutputs(l)
}

// Debug records debugging information. This won't do anything if the current
// module isn't beind debugged.
func (l Log) Debug(v ...interface{}) {
	if !l.hasDebug() {
		return
	}
	l.Msg = fmt.Sprint(v...)
	l.Level = LevelDbg
	Config.RunOutputs(l)
}

// Debugf records debugging information. This won't do anything if the current
// module isn't beind debugged.
func (l Log) Debugf(f string, v ...interface{}) {
	if !l.hasDebug() {
		return
	}
	l.Msg = fmt.Sprintf(f, v...)
	l.Level = LevelDbg
	Config.RunOutputs(l)
}

func (l Log) Trace(v ...interface{}) Log {
	l.Msg = fmt.Sprint(v...)
	l.Level = LevelTrace
	if l.hasDebug() {
		Config.RunOutputs(l)
		return l
	}

	l.Traces = append(l.Traces, Config.Format(l))
	return l
}

func (l Log) Tracef(f string, v ...interface{}) Log {
	l.Msg = fmt.Sprintf(f, v...)
	l.Level = LevelTrace
	if l.hasDebug() {
		Config.RunOutputs(l)
		return l
	}

	l.Traces = append(l.Traces, Config.Format(l))
	return l
}

// FieldsSince adds timing information recorded with Since as fields.
func (l Log) FieldsSince() Log { return l.Fields(l.sinceLog) }

// FieldsRequest adds information from a HTTP request as fields.
func (l Log) FieldsRequest(r *http.Request) Log {
	if r == nil {
		panic("zlog.FieldsRequest: *http.Request is nil")
	}

	return l.Fields(F{
		"http_method":     r.Method,
		"http_url":        r.URL.String(),
		"http_form":       r.Form.Encode(),
		"http_host":       r.Host,
		"http_user_agent": r.UserAgent(),
	})
}

// FieldsLocation records the caller location.
func (l Log) FieldsLocation() Log {
	if _, file, line, ok := runtime.Caller(2); ok {
		l = l.Fields(F{"location": fmt.Sprintf("%s:%d", filepath.Base(file), line)})
	}
	return l
}

func (l Log) hasDebug() bool {
	for _, m := range l.Modules {
		for _, d := range Config.Debug {
			if d == "all" || d == m {
				return true
			}
		}

		for _, d := range l.DebugModules {
			if d == "all" || d == m {
				return true
			}
		}
	}
	return false
}

var stderr io.Writer = os.Stderr // So we can swap it out in test.

// Since records the duration since the last Since() or Module() call with the
// given message.
//
// The result will be printed to stderr if this module is in the debug list. It
// can also be added to a Log with FieldsSince().
func (l Log) Since(msg string) Log {
	n := time.Now()
	if l.since.IsZero() {
		l.since = n
	}

	s := time.Since(l.since).Nanoseconds() / 1000000
	if l.sinceLog == nil {
		l.sinceLog = make(F)
	}
	l.sinceLog[msg] = fmt.Sprintf("%dms", s)
	if l.hasDebug() {
		fmt.Fprintf(stderr, "  %-16s %5dms  %s\n", strings.Join(l.Modules, ":"), s, msg)
	}

	l.since = n
	return l
}

// Recover from a panic.
//
// Any panics will be recover()'d and reported with Error():
//
//   go func() {
//       defer zlog.Recover()
//       // ... do work...
//   }()
//
// The first callback will be called before the Error() call, and can be used to
// modify the Log instance, for example to add fields:
//
//   defer zlog.Recover(func(l zlog.Log) zlog.Log {
//       return l.Fields(zlog.F{"id": id})
//   })
//
// Any other callbacks will be called after the Error() call. Modifying the Log
// instance has no real use.
func Recover(cb ...func(Log) Log) {
	r := recover()
	if r == nil {
		return
	}

	err, ok := r.(error)
	if !ok {
		err = fmt.Errorf("%v", r)
	}

	l := Module("panic")
	if len(cb) > 0 {
		l = cb[0](l)
	}

	err = fmt.Errorf("%s\n%s", err, debug.Stack())

	l.Error(err)

	if len(cb) > 1 {
		for i := range cb[1:] {
			l = cb[i](l)
		}
	}
}

// ProfileCPU writes a memory if the path is non-empty. This should be called on
// start and the returned function on end (e.g. defer):
//
//   func main() {
//       defer zlog.ProfileCPU("cpu.prof")()
//
//       // ..work..
//   }
func ProfileCPU(path string) func() {
	if path == "" {
		return func() {}
	}

	fp, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(fp)
	return pprof.StopCPUProfile
}

// ProfileHeap writes a memory if the path is non-empty. This is usually called
// just before the program exits:
//
//   func main() {
//       // ..work..
//
//       zlog.ProfileHeap("mem.prof")
//   }
func ProfileHeap(path string) {
	if path == "" {
		return
	}

	fp, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	pprof.WriteHeapProfile(fp)
	fp.Close()
}
