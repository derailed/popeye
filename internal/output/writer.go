package output

import (
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal/k8s/linter"
)

// Color ANSI palette (256!)
const (
	ColorOrangish = 220
	ColorOrange   = 202
	ColorGray     = 246
	ColorWhite    = 15
	ColorBlue     = 105
	ColorRed      = 196
	ColorCoolBlue = 99
	ColorAqua     = 123
)

// FontBold style
const (
	FontBold = 1
)

// Color tracks the output color.
type Color int

const outputWidth = 80

// Dump all errors output.
func Dump(l linter.Level, section string, issues ...linter.Issue) {
	for _, i := range issues {
		if i.Severity() >= l {
			Write(i.Severity(), "", i.Description())
		}
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
	fmt.Printf("%-10s %s%s\n", prefix, msg, emojiForLevel(l))
	return
}

// Colorize a string based on given color.
func Colorize(s string, c Color) string {
	return fmt.Sprintf("\033[38;5;%d;m%s\033[0m", c, s)
}

func colorForLevel(l linter.Level) Color {
	switch l {
	case linter.ErrorLevel:
		return ColorRed
	case linter.WarnLevel:
		return ColorOrangish
	case linter.InfoLevel:
		return ColorAqua
	default:
		return ColorGray
	}
}

func emojiForLevel(l linter.Level) string {
	switch l {
	case linter.ErrorLevel:
		return "💥"
	case linter.WarnLevel:
		return "️️⚠️"
	case linter.InfoLevel:
		return "🔊"
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
