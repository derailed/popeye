package output

import (
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal/k8s/linter"
)

// Defines basic ANSI colors.
const (
	ColorBlack Color = 30 + iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

// Brighter colors
const (
	ColorBriteBlack Color = 90 + iota
	ColorBriteRed
	ColorBriteGreen
	ColorBriteYellow
	ColorBriteBlue
	ColorBriteMagenta
	ColorBriteCyan
	ColorBriteWhite
)

// FontBold style
const (
	FontBold = 1
)

// Color tracks the output color.
type Color int

const outputWidth = 80

// Dump all errors output.
func Dump(section string, issues ...linter.Issue) {
	for _, i := range issues {
		Write(i.Severity(), "", i.Description())
	}
}

// Write a colorized message to stdout.
func Write(l linter.Level, prefix, msg string) {
	if prefix == "" {
		msg = Colorize(msg, colorForLevel(l))
		fmt.Printf("%s%s %s\n", strings.Repeat(" ", 13), "o", msg)
		return
	}

	dots := outputWidth - len(msg)
	dots -= 10 + 1
	msg = Colorize(msg+strings.Repeat(".", dots), colorForLevel(l))
	fmt.Printf("%10s %s%s\n", prefix, msg, emojiForLevel(l))
	return
}

// Colorize a string based on given color.
func Colorize(s string, c Color) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", c, s)
}

func colorForLevel(l linter.Level) Color {
	switch l {
	case linter.ErrorLevel:
		return ColorBriteRed
	case linter.WarnLevel:
		return ColorYellow
	case linter.InfoLevel:
		return ColorBlue
	default:
		return ColorBriteBlack
	}
}

func emojiForLevel(l linter.Level) string {
	switch l {
	case linter.ErrorLevel:
		return "💥"
	case linter.WarnLevel:
		return "️️⚠️"
	case linter.InfoLevel:
		return "🛠"
	default:
		return "✅"
	}
}

var Logo = []string{
	"           .-'-.      ",
	"        __|     `\\   ",
	"       `-,-`--._  `\\ ",
	" []   .->'  a   `|-'  ",
	"  `=/ (__/_      /    ",
	"    \\_,    `   _)    ",
	"      `----;  |       ",
}
