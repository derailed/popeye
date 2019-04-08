package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/derailed/popeye/internal/linter"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	// FontBold style
	FontBold = 1

	// Width denotes the maximum width of the sanitizer report.
	Width   = 100
	tabSize = 2
)

// Open begins a new report section.
func Open(w io.Writer, s string, t *Tally) {
	fmt.Fprintf(w, "\n%s", Colorize(s, ColorLighSlate))
	if t != nil && t.IsValid() {
		indent := Width - len(s) - t.Width() + 13
		fmt.Fprintf(w, "%s", strings.Repeat(" ", indent))
		t.Dump(w)
	}
	fmt.Fprintf(w, "\n%s", Colorize(strings.Repeat("┅", Width), ColorLighSlate))
	fmt.Fprintln(w)
}

// Close a report section.
func Close(w io.Writer) {
	fmt.Fprintln(w)
}

func lineBreaks(w io.Writer, s string, width int, color Color) {
	for i := 0; len(s) > width; i++ {
		fmt.Fprintln(w, Colorize(s[:width], color))
		s = s[width:]
	}
	if len(s) > 0 {
		fmt.Fprint(w, Colorize(s, color))
	}
	fmt.Fprintln(w)
}

// Error prints out error out.
func Error(w io.Writer, msg string, err error) {
	fmt.Fprintln(w)
	msg = msg + ": " + err.Error()
	width := Width - 3
	fmt.Fprintf(w, "💥 ")
	lineBreaks(w, msg, width, ColorRed)
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

	maxWidth := Width - tabSize*indent - 3
	msg = truncate(msg, maxWidth)
	if indent == 1 {
		dots := maxWidth - len(msg)
		msg = Colorize(msg, colorForLevel(l)) + Colorize(strings.Repeat(".", dots), ColorGray)
		fmt.Fprintf(w, "%s· %s%s\n", spacer, msg, EmojiForLevel(l))
		return
	}

	msg = Colorize(msg, colorForLevel(l))
	fmt.Fprintf(w, "%s%s %s\n", spacer, EmojiForLevel(l), msg)
}

// Truncate a string to the given l and suffix ellipsis if needed.
func truncate(str string, width int) string {
	return runewidth.Truncate(str, width, "...")
}
