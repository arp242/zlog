package log

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/teamwork/utils/errorutil"
)

func Test(t *testing.T) {
	n := time.Now()
	now = func() time.Time { return n }

	Config.StackFilter = errorutil.FilterPattern(errorutil.FilterTraceInclude, "testing")

	reTrace := regexp.MustCompile(`\t/.*?/testing\.go:\d+\n`)

	tests := []struct {
		in   func()
		want string
	}{
		{func() { Print("w00t") }, "w00t\n"},
		{func() { Printf("w00t %s", "x") }, "w00t x\n"},
		{func() { Error(errors.New("w00t")) }, "w00t\n\t\n\ttesting.tRunner\n\t\t/fake/testing.go:42\n\t"},
		{func() { Errorf("w00t %s", "x") }, "w00t x\n\t\n\ttesting.tRunner\n\t\t/fake/testing.go:42\n\t"},

		{func() { Module("test").Print("w00t") }, "test: w00t\n"},
		{func() { Module("test").Module("second").Print("w00t") }, "test: second: w00t\n"},
		{func() { Module("test").Error(errors.New("w00t")) }, "test: w00t\n\t\n\ttesting.tRunner\n\t\t/fake/testing.go:42\n\t"},

		{func() { Module("test").Fields(F{"k": "v"}).Print("w00t") }, "test: w00t k=v\n"},
		{func() { Module("test").Fields(F{"k": 3}).Print("w00t") }, "test: w00t k=3\n"},

		{func() { Module("test").Debug("w00t") }, ""},
		{func() { Debug("xxx").Module("test").Debug("w00t") }, ""},
		{func() { Debug("test").Module("test").Debug("w00t") }, "test: w00t\n"},

		{func() { Module("test").Trace("w00t") }, ""},
		{func() { Debug("test").Module("test").Trace("w00t") }, "test: w00t\n"},
		{func() { Module("test").Trace("w00t").Errorf("oh noes") }, "test: w00t\n" + n.Format(Config.FmtTime) + "test: oh noes\n\t\n\ttesting.tRunner\n\t\t/fake/testing.go:42\n\t"},
		{func() { Module("test").Trace("w00t").Print("print") }, "test: print\n"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			var buf bytes.Buffer
			Config.Output = func(l Log) { buf.WriteString(Config.Format(l)) }

			tt.in()
			out := buf.String()
			out = reTrace.ReplaceAllString(out, "\t/fake/testing.go:42\n")

			if tt.want != "" {
				tt.want = n.Format(Config.FmtTime) + tt.want
			}
			if out != tt.want {
				t.Errorf("\nout:  %q\nwant: %q\n", out, tt.want)
			}
		})
	}
}
