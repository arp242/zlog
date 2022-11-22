package zlog

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"zgo.at/zlog/internal/isatty"
)

var (
	enableColors = isatty.IsTerminal(os.Stdout.Fd())

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

// JSON strings aren't quoted in the output.
type JSON string

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
		width := 0
		for k := range l.Data {
			if l := len(k); l > width {
				width = l
			}
		}

		data := make([]string, len(l.Data))
		i := 0
		for k, v := range l.Data {
			vfmt := "%v"
			switch v.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint64:
				vfmt = "%d"
			case float32, float64:
				vfmt = "%f"
			case JSON:
				vfmt = "%s"
			case string, []byte, []rune:
				vfmt = "%q"
			case bool:
				vfmt = "%t"
			default:
				data[i] = fmt.Sprintf("%s = %v", k, v)
			}

			pad := strings.Repeat(" ", width-len(k))
			data[i] = fmt.Sprintf("%s%s = "+vfmt, k, pad, v)

			i++
		}

		sort.Strings(data) // Map order is random, so be predictable.

		//b.WriteString(" {")
		b.WriteString("\n\t")
		b.WriteString(strings.Join(data, "\n\t"))
		//b.WriteString("}")
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
