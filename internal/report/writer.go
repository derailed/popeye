// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/rules"
)

// Issue represents a linter issue.
type Issue interface {
	// MaxSeverity
	MaxSeverity(string) rules.Level
	Severity() rules.Level
	Description() string
	HasSubIssues() bool
	SubIssues() map[string][]Issue
}

const (
	// FontBold style
	FontBold = 1

	// Width denotes the maximum width of a report.
	Width = 100

	tabSize = 2
)

// ScanReport represents a scan report.
type ScanReport struct {
	io.Writer

	jurassicMode bool
}

//

// New returns a new instance.
func New(w io.Writer, jurassic bool) *ScanReport {
	return &ScanReport{
		Writer:       w,
		jurassicMode: jurassic,
	}
}

// Open begins a new report section.
func (s *ScanReport) Open(msg string, t *Tally) {
	fmt.Fprintf(s, "\n%s", s.Color(msg, ColorLighSlate))
	if t != nil && t.IsValid() {
		out := t.Dump(s)
		spacer := 12
		if s.jurassicMode {
			spacer = 2
		}
		indent := Width - len(msg) - utf8.RuneCountInString(out) + spacer
		fmt.Fprintf(s, "%s", strings.Repeat(" ", indent))
		fmt.Fprintf(s, "%s", out)
	}
	titleSeparator := "┅"
	if s.jurassicMode {
		titleSeparator = "="
	}
	fmt.Fprintf(s, "\n%s", s.Color(strings.Repeat(titleSeparator, Width+1), ColorLighSlate))
	fmt.Fprintln(s)
}

// Close a report section.
func (s *ScanReport) Close() {
	fmt.Fprintln(s)
}

func (s *ScanReport) lineBreaks(msg string, width int, color Color) {
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
func (s *ScanReport) Error(msg string, err error) {
	fmt.Fprintln(s)
	msg = msg + ": " + err.Error()
	width := Width - 3
	fmt.Fprintf(s, "💥 ")
	s.lineBreaks(msg, width, ColorRed)
}

// Comment writes a comment line.
func (s *ScanReport) Comment(msg string) {
	fmt.Fprint(s, "  · "+msg+"\n")
}

// Dump all errors to output.
func (s *ScanReport) Dump(l rules.Level, ii issues.Issues) {
	groups := ii.Group()
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, group := range keys {
		sev := groups[group].MaxSeverity()
		if sev < l {
			continue
		}
		if group != issues.Root {
			s.write(containerLevel, 2, group)
		}
		for _, i := range groups[group] {
			if i.Level < l {
				continue
			}
			if i.Group == issues.Root {
				s.write(i.Level, 2, i.Message+".")
				continue
			}
			s.write(i.Level, 3, i.Message+".")
		}
	}
}

// Print a colorized message.
func (s *ScanReport) Print(l rules.Level, indent int, msg string) {
	s.write(l, indent, msg)
}

// Write a colorized message to stdout.
func (s *ScanReport) write(l rules.Level, indent int, msg string) {
	if msg == "" || msg == "." {
		return
	}

	spacer, emoji := strings.Repeat(" ", tabSize*indent), EmojiForLevel(l, s.jurassicMode)

	extra := 1
	if s.jurassicMode {
		extra--
	}
	maxWidth := Width - tabSize*indent - utf8.RuneCountInString(emoji) - 1
	msg = formatLine(msg, indent, maxWidth)
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
	if emoji == "" {
		fmt.Fprintf(s, "%s%s\n", spacer, msg)
	} else {
		fmt.Fprintf(s, "%s%s %s\n", spacer, emoji, msg)
	}
}

// Color or not this message by inject ansi colors.
func (s *ScanReport) Color(msg string, c Color) string {
	if s.jurassicMode {
		return msg
	}
	return Colorize(msg, c)
}

// ----------------------------------------------------------------------------
// Helpers...

// Truncate a string to the given l and suffix ellipsis if needed.
func formatLine(str string, indent, width int) string {
	if len(str) <= width {
		return str
	}

	tokens, length := strings.Split(str, " "), 0
	var lines []string
	for len(tokens) > 0 {
		var line string
		if len(lines) > 0 {
			line = strings.Repeat("  ", tabSize*indent)
			length += tabSize * indent * 2
		}
		for _, t := range tokens {
			l := len(t) + 1
			if length+l > width {
				lines = append(lines, line)
				spacer := strings.Repeat(" ", tabSize*indent+3)
				line = spacer + t + " "
				length = len(spacer) + l
			} else {
				line += t + " "
				length += l
			}
			tokens = tokens[1:]
		}
		lines = append(lines, line)
		length = 0
	}
	return strings.Join(lines, "\n")
}
