// Package zlog is a logging library.
package zlog // import "zgo.at/zlog"

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/teamwork/utils/errorutil"
)

// ConfigT is the configuration struct.
type ConfigT struct {
	// Time/date format as accepted by time.Format().
	//
	// The default is to just print the time, which works well in development.
	// For production you probably want to use time.RFC3339 or some such.
	//
	// This is used in the standard format() function, not not elsewhere.
	FmtTime string

	// Format function; this takes a Log entry and formats it for output.
	//
	// This can also return JSON or something else, whatever makes sense for
	// your Output function.
	Format func(Log) string

	// Output a Log entry; this is expected to call Format() and then do
	// *something* with the result.
	//
	// The default is to print to stderr for errors, and stdout for everything
	// else.
	Output func(Log)

	// Filter stack traces, only used if the error is a github.com/pkg/errors
	// with a stack trace.
	// Useful to filter out HTTP middleware and other useless stuff.
	StackFilter *errorutil.Patterns

	// Global debug modules; always print debug information for these.
	Debug []string
}

// Config for this package.
var Config ConfigT

func init() {
	Config = ConfigT{
		FmtTime: "15:04:05 ",
		Format:  format,
		Output:  output,
	}
}

const (
	//levelPrint = 0
	levelErr   = 1
	levelDbg   = 2
	levelTrace = 3
)

var now = func() time.Time { return time.Now() }

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
	}

	// F are log fields.
	F map[string]interface{}
)

func Module(m string) Log   { return Log{Modules: []string{m}} }
func Fields(f F) Log        { return Log{Data: f} }
func Debug(m ...string) Log { return Log{DebugModules: m} }

func Print(v ...interface{})            { Log{}.Print(v...) }
func Printf(f string, v ...interface{}) { Log{}.Printf(f, v...) }
func Error(err error)                   { Log{}.Error(err) }
func Errorf(f string, v ...interface{}) { Log{}.Errorf(f, v...) }

func (l Log) ResetTrace()                 { l.Traces = []string{} }
func (l Log) Context(ctx context.Context) { l.Ctx = ctx }

func (l Log) Module(m string) Log {
	l.Modules = append(l.Modules, m)
	return l
}
func (l Log) Fields(f F) Log {
	l.Data = f
	return l
}

func (l Log) Print(v ...interface{}) {
	l.Msg = fmt.Sprint(v...)
	Config.Output(l)
}
func (l Log) Printf(f string, v ...interface{}) {
	l.Msg = fmt.Sprintf(f, v...)
	Config.Output(l)
}
func (l Log) Error(err error) {
	l.Err = err
	l.Level = levelErr
	Config.Output(l)
}
func (l Log) Errorf(f string, v ...interface{}) {
	l.Err = fmt.Errorf(f, v...)
	l.Level = levelErr
	Config.Output(l)
}

func (l Log) Debug(v ...interface{}) {
	if !l.hasDebug() {
		return
	}
	l.Level = levelDbg
	l.Print(v...)
}
func (l Log) Debugf(f string, v ...interface{}) {
	if !l.hasDebug() {
		return
	}
	l.Level = levelDbg
	l.Printf(f, v...)
}

func (l Log) Trace(v ...interface{}) Log {
	l.Msg = fmt.Sprint(v...)
	l.Level = levelTrace
	if l.hasDebug() {
		l.Print(v...)
		return l
	}

	l.Traces = append(l.Traces, Config.Format(l))
	return l
}

func (l Log) Tracef(f string, v ...interface{}) Log {
	l.Msg = fmt.Sprintf(f, v...)
	l.Level = levelTrace
	if l.hasDebug() {
		l.Printf(f, v...)
		return l
	}

	l.Traces = append(l.Traces, Config.Format(l))
	return l
}

// TODO: we could store as map[string]struct{} internally so it's a bit faster.
// Not sure if that's actually worth it?
func (l Log) hasDebug() bool {
	for _, m := range l.Modules {
		for _, d := range Config.Debug {
			if d == m {
				return true
			}
		}

		for _, d := range l.DebugModules {
			if d == m {
				return true
			}
		}
	}
	return false
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}
