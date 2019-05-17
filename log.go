// Package log is a logging library.
package log // import "arp242.net/log"

import (
	"context"
	"fmt"
	"os"
	"strings"
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
	FmtTime string

	// Format function; this takes a Log entry and formats it for output.
	//
	// This can also return JSON or something else, whatever makes sense for
	// your Output function.
	Format func(Log) string

	// Output a Log entry; this is expected to call Format() and then do
	// *something* with it.
	//
	// The default is to print to stderr for errors, and stdout for everything
	// else. It can also
	Output func(Log)

	// Filter stack traces, only used if the error is a github.com/pkg/errors
	// with a stack trace.
	// Useful to filter out HTTP middlewares and other useless stuff.
	StackFilter *errorutil.Patterns
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
	Log struct {
		msg     string   // Log message.
		err     error    // Original error, in case of errors.
		level   int      // 0: print, 1: err, 2: debug, 3: trace
		modules []string // Modules added to the logger.
		fields  F        // Fields added to the logger.
		debug   []string // List of modules to debug.
		traces  []string // Traces added to the logger.
		ctx     context.Context
	}

	F map[string]interface{}
)

func Module(m string) Log   { return Log{modules: []string{m}} }
func Fields(f F) Log        { return Log{fields: f} }
func Debug(m ...string) Log { return Log{debug: m} }

func Print(v ...interface{})            { Log{}.Print(v...) }
func Printf(f string, v ...interface{}) { Log{}.Printf(f, v...) }
func Error(err error)                   { Log{}.Error(err) }
func Errorf(f string, v ...interface{}) { Log{}.Errorf(f, v...) }

func (l Log) ResetTrace()                 { l.traces = []string{} }
func (l Log) Context(ctx context.Context) { l.ctx = ctx }

func (l Log) Module(m string) Log {
	l.modules = append(l.modules, m)
	return l
}
func (l Log) Fields(f F) Log {
	l.fields = f
	return l
}

func (l Log) Print(v ...interface{}) {
	l.msg = fmt.Sprint(v...)
	Config.Output(l)
}
func (l Log) Printf(f string, v ...interface{}) {
	l.msg = fmt.Sprintf(f, v...)
	Config.Output(l)
}
func (l Log) Error(err error) {
	l.msg = fmt.Sprintf("%s", err)
	l.level = levelErr
	Config.Output(l)
}
func (l Log) Errorf(f string, v ...interface{}) {
	l.msg = fmt.Sprintf(f, v...)
	l.level = levelErr
	Config.Output(l)
}

func (l Log) Debug(v ...interface{}) {
	if !l.hasDebug() {
		return
	}
	l.level = levelDbg
	l.Print(v...)
}
func (l Log) Debugf(f string, v ...interface{}) {
	if !l.hasDebug() {
		return
	}
	l.level = levelDbg
	l.Printf(f, v...)
}

func (l Log) Trace(v ...interface{}) Log {
	l.msg = fmt.Sprint(v...)
	l.level = levelTrace
	if l.hasDebug() {
		l.Print(v...)
		return l
	}

	l.traces = append(l.traces, Config.Format(l))
	return l
}

func (l Log) Tracef(f string, v ...interface{}) Log {
	l.msg = fmt.Sprintf(f, v...)
	l.level = levelTrace
	if l.hasDebug() {
		l.Printf(f, v...)
		return l
	}

	l.traces = append(l.traces, Config.Format(l))
	return l
}

func (l Log) hasDebug() bool {
	for _, d := range l.debug {
		for _, m := range l.modules {
			if d == m {
				return true
			}
		}
	}
	return false
}

func format(l Log) string {
	b := &strings.Builder{}

	// Write any existing trace logs on error.
	if l.level == levelErr {
		for _, t := range l.traces {
			b.Write([]byte(t))
		}
	}

	b.WriteString(now().Format(Config.FmtTime))
	if len(l.modules) > 0 {
		b.WriteString(strings.Join(l.modules, ": "))
		b.WriteString(": ")
	}
	b.WriteString(l.msg)

	if len(l.fields) > 0 {
		b.WriteString(" ")
		for k, v := range l.fields {
			fmt.Fprintf(b, "%s=%v", k, v)
		}
	}

	b.WriteString("\n")

	// TODO: also support new error interface in Go 1.13
	if l.level == levelErr {
		if l.err == nil {
			l.err = errors.WithStack(errors.New(""))
		} else if _, ok := l.err.(stackTracer); !ok {
			l.err = errors.WithStack(l.err)
		}

		if Config.StackFilter != nil {
			l.err = errorutil.FilterTrace(l.err, Config.StackFilter)
		}

		st := l.err.(stackTracer)
		b.WriteString(strings.Replace(fmt.Sprintf("\t%+v\n", st.StackTrace()), "\n", "\n\t", -1))
	}

	return b.String()
}

func output(l Log) {
	out := os.Stdout
	if l.level == levelErr {
		out = os.Stderr
	}
	fmt.Fprintln(out, Config.Format(l))
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}
