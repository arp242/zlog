package zlog

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/teamwork/utils/errorutil"
)

func Test(t *testing.T) {
	n := time.Now()
	now = func() time.Time { return n }
	enableColors = false

	Config.StackFilter = errorutil.FilterPattern(errorutil.FilterTraceInclude, "testing")

	reTrace := regexp.MustCompile(`\t/.*?/testing\.go:\d+`)

	tests := []struct {
		in   func()
		want string
	}{
		{func() { Print("w00t") }, "INFO: w00t"},
		{func() { Printf("w00t %s", "x") }, "INFO: w00t x"},
		{func() { Error(errors.New("w00t")) }, "ERROR: w00t\n\ttesting.tRunner\n\t\t/fake/testing.go:42"},
		{func() { Errorf("w00t %s", "x") }, "ERROR: w00t x\n\ttesting.tRunner\n\t\t/fake/testing.go:42"},

		{func() { Module("test").Print("w00t") }, "test: INFO: w00t"},
		{func() { Module("test").Module("second").Print("w00t") }, "test: second: INFO: w00t"},
		{func() { Module("test").Error(errors.New("w00t")) }, "test: ERROR: w00t\n\ttesting.tRunner\n\t\t/fake/testing.go:42"},

		{func() { Module("test").Fields(F{"k": "v"}).Print("w00t") }, "test: INFO: w00t k=v"},
		{func() { Module("test").Fields(F{"k": 3}).Print("w00t") }, "test: INFO: w00t k=3"},

		{func() { Module("test").Debug("w00t") }, ""},
		{func() { Debug("xxx").Module("test").Debug("w00t") }, ""},
		{func() { Debug("test").Module("test").Debug("w00t") }, "test: DEBUG: w00t"},
		{func() { Debug("test").Module("test").Debugf("w00t %d", 42) }, "test: DEBUG: w00t 42"},

		{func() { Module("test").Trace("w00t") }, ""},
		{func() { Debug("test").Module("test").Trace("w00t") }, "test: TRACE: w00t"},
		{func() { Module("test").Tracef("w00t %d", 42).Errorf("oh noes") }, "test: TRACE: w00t 42\n" + n.Format(Config.FmtTime) + "test: ERROR: oh noes\n\ttesting.tRunner\n\t\t/fake/testing.go:42"},
		{func() { Module("test").Trace("w00t").Print("print") }, "test: INFO: print"},
		{func() { Module("test").Tracef("w00t").Print("print") }, "test: INFO: print"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			var buf bytes.Buffer
			Config.Output = func(l Log) { buf.WriteString(Config.Format(l)) }

			tt.in()
			out := buf.String()
			out = reTrace.ReplaceAllString(out, "\t/fake/testing.go:42")

			if tt.want != "" {
				tt.want = n.Format(Config.FmtTime) + tt.want
			}
			if out != tt.want {
				t.Errorf("\nout:  %q\nwant: %q\n", out, tt.want)
			}
		})
	}
}

func BenchmarkPrint(b *testing.B) {
	text := strings.Repeat("Hello, world, it's a sentences!\n", 4)
	for n := 0; n < b.N; n++ {
		Print(text)
	}
}

func BenchmarkFields(b *testing.B) {
	l := Module("bench").Fields(F{
		"a": "b",
		"c": "d",
	})
	text := strings.Repeat("Hello, world, it's a sentences!\n", 4)
	for n := 0; n < b.N; n++ {
		l.Print(text)
	}
}
