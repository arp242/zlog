package zlog

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	n := time.Now()
	now = func() time.Time { return n }
	enableColors = false

	tests := []struct {
		in   func()
		want string
	}{
		// TODO: get line nr. instead of hard-coding it here.
		{func() { FieldsLocation().Print("print") }, "INFO: print {location=\"zlog_test.go:25\"}"},

		{func() { Print("w00t") }, "INFO: w00t"},
		{func() { Printf("w00t %s", "x") }, "INFO: w00t x"},
		{func() { Error(errors.New("w00t")) }, "ERROR: w00t"},
		{func() { Errorf("w00t %s", "x") }, "ERROR: w00t x"},

		{func() { Module("test").Print("w00t") }, "test: INFO: w00t"},
		{func() { Module("test").Module("second").Print("w00t") }, "test: second: INFO: w00t"},
		{func() { Module("test").Error(errors.New("w00t")) }, "test: ERROR: w00t"},

		{func() { Module("test").Fields(F{"k": "v"}).Print("w00t") }, "test: INFO: w00t {k=\"v\"}"},
		{func() { Module("test").Fields(F{"k": 3}).Print("w00t") }, "test: INFO: w00t {k=3}"},
		{func() { Module("test").Fields(F{"k": 3}).Field("k", "s").Print("w00t") }, "test: INFO: w00t {k=\"s\"}"},
		{func() { Module("test").Field("k", 404).Print("w00t") }, "test: INFO: w00t {k=404}"},

		{func() { Module("test").Debug("w00t") }, ""},
		{func() { SetDebug("xxx").Module("test").Debug("w00t") }, ""},
		{func() { SetDebug("test").Module("test").Debug("w00t") }, "test: DEBUG: w00t"},
		{func() { Module("test").SetDebug("test").Debug("w00t") }, "test: DEBUG: w00t"},
		{func() { SetDebug("test").Module("test").Debugf("w00t %d", 42) }, "test: DEBUG: w00t 42"},

		{func() { Module("test").Trace("w00t") }, ""},
		{func() { SetDebug("test").Module("test").Trace("w00t") }, "test: TRACE: w00t"},
		{func() { Module("test").Tracef("w00t %d", 42).Errorf("oh noes") }, "test: TRACE: w00t 42\n" + n.Format(Config.FmtTime) + "test: ERROR: oh noes"},
		{func() { Module("test").Trace("w00t").Print("print") }, "test: INFO: print"},
		{func() { Module("test").Tracef("w00t").Print("print") }, "test: INFO: print"},

		{func() {
			r, _ := http.NewRequest("PUT", "/path?k=v&a=b", nil)
			FieldsRequest(r).Error(errors.New("w00t"))
		}, `ERROR: w00t {http_form="" http_headers="" http_host="" http_method="PUT" http_url="/path?k=v&a=b"}`},
		{func() {
			r, _ := http.NewRequest("PUT", "/path?k=v&a=b", nil)
			r.Header.Set("User-Agent", "x")
			FieldsRequest(r).Print("w00t")
		}, `INFO: w00t {http_form="" http_headers="User-Agent: x" http_host="" http_method="PUT" http_url="/path?k=v&a=b"}`},
		{func() {
			r, _ := http.NewRequest("PUT", "/path?k=v&a=b", nil)
			r.Header.Set("User-Agent", "x")
			r.Header.Set("Host", "foo.com")
			r.Header.Set("Other-Head", `"quoted"`)
			r.Header.Set("Multi", `1`)
			r.Header.Set("Multi", `2`)
			FieldsRequest(r).Print("w00t")
		}, `INFO: w00t {http_form="" http_headers="Host: foo.com · Multi: 2 · Other-Head: \"quoted\" · User-Agent: x" http_host="" http_method="PUT" http_url="/path?k=v&a=b"}`},

		{func() {
			Config.SetDebug("all")
			Module("c").Debug("xxx")
			Config.SetDebug("")
			Module("c").Debug("xxx")
		}, "c: DEBUG: xxx"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			var buf bytes.Buffer
			var lock sync.Mutex
			Config.Outputs = []OutputFunc{
				func(l Log) {
					lock.Lock()
					buf.WriteString(Config.Format(l))
					lock.Unlock()
				},
			}

			tt.in()
			out := buf.String()

			if tt.want != "" {
				tt.want = n.Format(Config.FmtTime) + tt.want
			}
			if out != tt.want {
				t.Errorf("\nout:  %s\nwant: %s\n", out, tt.want)
			}
		})
	}
}

func TestFields(t *testing.T) {
	n := time.Now()
	now = func() time.Time { return n }
	enableColors = false

	var (
		buf  bytes.Buffer
		lock sync.Mutex
	)
	Config.Outputs = []OutputFunc{
		func(l Log) {
			lock.Lock()
			buf.WriteString(Config.Format(l))
			lock.Unlock()
		},
	}

	//o := 1
	Fields(F{
		"x": struct {
			i int64
			//p *int // TODO: displays ptr address
		}{123}, //&o},
		"y": map[string]string{"X": "Y"},
		"z": 690123,
		"b": []byte("aa"),
	}).Print("p")

	out := buf.String()
	want := n.Format(Config.FmtTime) + `INFO: p {b="aa" x={123} y=map[X:Y] z=690123}`
	if out != want {
		t.Errorf("\nout:  %q\nwant: %q", out, want)
	}
}

func TestSince(t *testing.T) {
	tests := []struct {
		in   func()
		want string
	}{
		{func() { Module("test").Since("xxx") }, ""},
		{func() { SetDebug("test").Module("test").Since("xxx") }, "  test                 0ms  xxx\n"},
		{func() {
			l := SetDebug("test").Module("test").Since("xxx")
			time.Sleep(2 * time.Millisecond)
			l.Since("yyy")
			time.Sleep(4 * time.Millisecond)
			l.Since("zzz")
		}, "  test                 0ms  xxx\n  test                 2ms  yyy\n  test                 6ms  zzz\n"},
		{func() {
			l := SetDebug("test").Module("test").Since("xxx")
			time.Sleep(2 * time.Millisecond)
			l = l.Since("yyy")
			time.Sleep(4 * time.Millisecond)
			l.Since("zzz")
		}, "  test                 0ms  xxx\n  test                 2ms  yyy\n  test                 4ms  zzz\n"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			var buf bytes.Buffer
			stderr = &buf
			defer func() { stderr = os.Stderr }()

			tt.in()
			out := buf.String()
			if out != tt.want {
				t.Errorf("\nout:  %q\nwant: %q\n", out, tt.want)
			}
		})
	}

	t.Run("SinceLog", func(t *testing.T) {
		var buf bytes.Buffer
		var lock sync.Mutex
		Config.Outputs = []OutputFunc{
			func(l Log) {
				lock.Lock()
				buf.WriteString(Config.Format(l))
				lock.Unlock()
			},
		}

		l := Module("test").Since("xxx").Fields(F{"1": 2})
		time.Sleep(2 * time.Millisecond)
		l = l.Since("yyy")
		time.Sleep(4 * time.Millisecond)
		l.Since("zzz")
		l.FieldsSince().Print("msg")

		out := buf.String()
		out = out[strings.Index(out, ":")+7:]
		want := "test: INFO: msg {1=2 xxx=\"0ms\" yyy=\"2ms\" zzz=\"4ms\"}"
		if out != want {
			t.Errorf("\nout:  %q\nwant: %q\n", out, want)
		}
	})
}

// TODO: expand test (i.e. test that it works beyond running).
func TestRecover(t *testing.T) {
	go func() {
		defer Recover()
	}()

	go func() {
		defer Recover()
		panic("oh noes")
	}()

	go func() {
		defer Recover(func(l Log) Log {
			return l.Fields(F{"a": "b"})
		},
			func(l Log) Log {
				fmt.Println("after")
				return l
			})
		panic("oh noes")
	}()
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
