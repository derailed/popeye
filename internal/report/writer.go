package report

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/derailed/popeye/internal/linter"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/onsi/ginkgo/reporters/stenographer/support/go-isatty"
	"github.com/rs/zerolog/log"
)

const (
	// FontBold style
	FontBold = 1

	// Width denotes the maximum width of the sanitizer report.
	Width   = 100
	tabSize = 2
)

// Sanitizer represents a sanitizer report.
type Sanitizer struct {
	io.Writer

	jurassicMode bool
}

//

// NewSanitizer returns a new sanitizer report writer.
func NewSanitizer(w io.Writer, fd uintptr, jurassic *bool) *Sanitizer {
	s := Sanitizer{Writer: w, jurassicMode: jurassicTerm(fd)}

	if jurassic != nil {
		s.jurassicMode = *jurassic
	}
	log.Debug().Msgf("Nehenderterm mode activated? %t", s.jurassicMode)

	return &s
}

// Open begins a new report section.
func (s *Sanitizer) Open(msg string, t *Tally) {
	fmt.Fprintf(s, "\n%s", s.Color(msg, ColorLighSlate))
	if t != nil && t.IsValid() {
		out := t.Dump(s)
		spacer := 13
		if s.jurassicMode {
			spacer = 1
		}
		indent := Width - len(msg) - utf8.RuneCountInString(out) + spacer
		fmt.Fprintf(s, "%s", strings.Repeat(" ", indent))
		fmt.Fprintf(s, out)
	}
	fmt.Fprintf(s, "\n%s", s.Color(strings.Repeat("┅", Width), ColorLighSlate))
	fmt.Fprintln(s)
}

// Close a report section.
func (s *Sanitizer) Close() {
	fmt.Fprintln(s)
}

func (s *Sanitizer) lineBreaks(msg string, width int, color Color) {
	for i := 0; len(msg) > width; i++ {
		fmt.Fprintln(s, s.Color(msg[:width], color))
		msg = msg[width:]
	}
	if len(msg) > 0 {
		fmt.Fprint(s, s.Color(msg, color))
	}
	fmt.Fprintln(s)
}

// Error prints out error out.
func (s *Sanitizer) Error(msg string, err error) {
	fmt.Fprintln(s)
	msg = msg + ": " + err.Error()
	width := Width - 3
	fmt.Fprintf(s, "💥 ")
	s.lineBreaks(msg, width, ColorRed)
}

// Comment writes a comment line.
func (s *Sanitizer) Comment(msg string) {
	fmt.Fprintf(s, "  · "+msg+"\n")
}

// Dump all errors to output.
func (s *Sanitizer) Dump(l linter.Level, issues ...linter.Issue) {
	var current string
	for _, i := range issues {
		if i.Severity() >= l {
			tokens := strings.Split(i.Description(), linter.Delimiter)
			if len(tokens) == 1 {
				s.write(i.Severity(), 2, i.Description()+".")
			} else {
				if current != tokens[0] {
					s.write(containerLevel, 2, tokens[0])
					current = tokens[0]
				}
				s.write(i.Severity(), 3, tokens[1]+".")
			}
		}
	}
}

// Print a colorized message.
func (s *Sanitizer) Print(l linter.Level, indent int, msg string) {
	s.write(l, indent, msg)
}

// Write a colorized message to stdout.
func (s *Sanitizer) write(l linter.Level, indent int, msg string) {
	spacer, emoji := strings.Repeat(" ", tabSize*indent), s.EmojiForLevel(l)

	extra := 1
	if s.jurassicMode {
		extra--
	}
	maxWidth := Width - tabSize*indent - utf8.RuneCountInString(emoji) - 1
	msg = truncate(msg, maxWidth)
	if indent == 1 {
		dots := maxWidth - len(msg) - extra
		if dots < 0 {
			dots = 0
		}
		msg = s.Color(msg, colorForLevel(l)) + s.Color(strings.Repeat(".", dots), ColorGray)
		fmt.Fprintf(s, "%s· %s%s\n", spacer, msg, emoji)
		return
	}

	msg = s.Color(msg, colorForLevel(l))
	fmt.Fprintf(s, "%s%s %s\n", spacer, emoji, msg)
}

// Color or not this message by inject ansi colors.
func (s *Sanitizer) Color(msg string, c Color) string {
	if s.jurassicMode {
		return msg
	}
	return Colorize(msg, c)
}

// ----------------------------------------------------------------------------
// Helpers...

// Truncate a string to the given l and suffix ellipsis if needed.
func truncate(str string, width int) string {
	return runewidth.Truncate(str, width, "...")
}

// Check terminal specs, returns true if lame term is effect. False otherwise.
func jurassicTerm(fd uintptr) bool {
	term := os.Getenv("TERM")
	if !strings.Contains(term, "color") {
		return false
	}

	return !isatty.IsTerminal(fd)
}
