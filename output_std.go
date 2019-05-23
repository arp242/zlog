package zlog

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/teamwork/utils/errorutil"
)

func format(l Log) string {
	b := &strings.Builder{}

	// Write any existing trace logs on error.
	if l.Level == levelErr {
		for _, t := range l.Traces {
			b.Write([]byte(t))
		}
	}

	b.WriteString(now().Format(Config.FmtTime))
	if len(l.Modules) > 0 {
		b.WriteString(strings.Join(l.Modules, ": "))
		b.WriteString(": ")
	}
	if l.Err != nil {
		b.WriteString(l.Err.Error())
	} else {
		b.WriteString(l.Msg)
	}

	if len(l.Data) > 0 {
		b.WriteString(" ")
		for k, v := range l.Data {
			fmt.Fprintf(b, "%s=%v", k, v)
		}
	}

	b.WriteString("\n")

	// TODO: also support new error interface in Go 1.13
	if l.Level == levelErr {
		if l.Err == nil {
			l.Err = errors.WithStack(errors.New(""))
		} else if _, ok := l.Err.(stackTracer); !ok {
			l.Err = errors.WithStack(l.Err)
		}

		if Config.StackFilter != nil {
			l.Err = errorutil.FilterTrace(l.Err, Config.StackFilter)
		}

		st := l.Err.(stackTracer)
		b.WriteString(strings.Replace(fmt.Sprintf("\t%+v\n", st.StackTrace()), "\n", "\n\t", -1))
	}

	return b.String()
}

func output(l Log) {
	out := os.Stdout
	if l.Level == levelErr {
		out = os.Stderr
	}
	fmt.Fprintln(out, Config.Format(l))
}
