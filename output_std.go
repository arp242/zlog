package log

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
