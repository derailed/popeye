package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/derailed/popeye/internal/linter"
)

const (
	// FontBold style
	FontBold = 1

	reportWidth = 80
	tabSize     = 2
)

// Open begins a new report section.
func Open(w io.Writer, s string, t *Tally) {
	fmt.Fprintf(w, "\n%s", Colorize(s, ColorLighSlate))
	if t != nil && t.IsValid() {
		indent := reportWidth - len(s) - t.Width() + 13
		fmt.Fprintf(w, "%s", strings.Repeat(" ", indent))
		t.Dump(w)
	}
	fmt.Fprintf(w, "\n%s", Colorize(strings.Repeat("┅", 80), ColorLighSlate))
	fmt.Fprintln(w)
}

// Close a report section.
func Close(w io.Writer) {
	fmt.Fprintln(w)
}

// Error prints out error out.
func Error(w io.Writer, msg string, err error) {
	fmt.Fprintln(w)
	msg = msg + ": " + err.Error()
	buff := make([]string, 0, len(msg)%reportWidth)
	width := reportWidth - 3
	for i := 0; len(msg) > width; i += width {
		buff = append(buff, msg[i:i+width])
		msg = msg[i+width:]
	}
	buff = append(buff, msg)
	fmt.Fprintf(w, "💥 "+Colorize(strings.Join(buff, "\n"), ColorRed))
	fmt.Fprintln(w)
}

// Comment writes a comment line.
func Comment(w io.Writer, msg string) {
	fmt.Fprintf(w, "  · "+msg+"\n")
}

// Dump all errors to output.
func Dump(w io.Writer, l linter.Level, issues ...linter.Issue) {
	var current string
	for _, i := range issues {
		if i.Severity() >= l {
			tokens := strings.Split(i.Description(), linter.Delimiter)
			if len(tokens) == 1 {
				Write(w, i.Severity(), 2, i.Description()+".")
			} else {
				if current != tokens[0] {
					Write(w, containerLevel, 2, tokens[0])
					current = tokens[0]
				}
				Write(w, i.Severity(), 3, tokens[1]+".")
			}
		}
	}
}

// Write a colorized message to stdout.
func Write(w io.Writer, l linter.Level, indent int, msg string) {
	spacer := strings.Repeat(" ", tabSize*indent)

	if indent == 1 {
		dots := reportWidth - len(msg) - tabSize*indent - 3
		msg = Colorize(msg, colorForLevel(l)) + Colorize(strings.Repeat(".", dots), ColorGray)
		fmt.Fprintf(w, "%s· %s%s\n", spacer, msg, EmojiForLevel(l))
		return
	}

	msg = Colorize(msg, colorForLevel(l))
	fmt.Fprintf(w, "%s%s %s\n", spacer, EmojiForLevel(l), msg)
}
