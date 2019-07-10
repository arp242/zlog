package zlog

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/teamwork/utils/errorutil"
)

var (
	enableColors = true

	// Fill two spaces with a background colour. Don't colourize the full text, as
	// this is more readable across different colour scheme choices.
	colors = map[int]string{
		LevelInfo:  "\x1b[48;5;12m  \x1b[0m ",  // Blue
		LevelErr:   "\x1b[48;5;9m  \x1b[0m ",   // Red
		LevelDbg:   "\x1b[48;5;247m  \x1b[0m ", // Grey
		LevelTrace: "\x1b[48;5;247m  \x1b[0m ", // Grey
	}

	messages = map[int]string{
		LevelInfo:  "INFO: ",
		LevelErr:   "ERROR: ",
		LevelDbg:   "DEBUG: ",
		LevelTrace: "TRACE: ",
	}
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func format(l Log) string {
	b := &strings.Builder{}

	// Write any existing trace logs on error.
	if l.Level == LevelErr {
		for _, t := range l.Traces {
			b.Write([]byte(t + "\n"))
		}
	}

	if enableColors {
		b.WriteString(colors[l.Level])
	}

	b.WriteString(now().Format(Config.FmtTime))
	if len(l.Modules) > 0 {
		b.WriteString(strings.Join(l.Modules, ": "))
		b.WriteString(": ")
	}

	b.WriteString(messages[l.Level])

	if l.Err != nil {
		b.WriteString(l.Err.Error())
	} else {
		b.WriteString(l.Msg)
	}

	if len(l.Data) > 0 {
		b.WriteString("\n\t{")
		first := true
		for k, v := range l.Data {
			if !first {
				b.WriteString(" ")
			} else {
				first = false
			}
			fmt.Fprintf(b, "%s=%q", k, v)
		}
		b.WriteString("}")
	}

	// TODO: also support new error interface in Go 1.13
	if l.Level == LevelErr {
		if l.Err == nil {
			l.Err = errors.WithStack(errors.New(""))
		} else if _, ok := l.Err.(stackTracer); !ok {
			l.Err = errors.WithStack(l.Err)
		}

		if Config.StackFilter != nil {
			l.Err = errorutil.FilterTrace(l.Err, Config.StackFilter)
		}

		st := l.Err.(stackTracer)
		trace := strings.TrimSpace(fmt.Sprintf("%+v", st.StackTrace()))
		indent := "\t" // TODO: would prefer to align with message using spaces.
		trace = indent + strings.Replace(trace, "\n", "\n"+indent, -1)
		//b.WriteString(strings.Replace(fmt.Sprintf("\t%+v\n", st.StackTrace()), "\n", "\n\t", -1))
		b.WriteString("\n")
		b.WriteString(trace)
	}

	return b.String()
}

func output(l Log) {
	out := os.Stdout
	if l.Level == LevelErr {
		out = os.Stderr
	}
	fmt.Fprintln(out, Config.Format(l))
}
