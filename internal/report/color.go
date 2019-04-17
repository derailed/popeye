package report

import (
	"strconv"

	"github.com/derailed/popeye/internal/linter"
)

// Color ANSI palette (256!)
const (
	ColorOrangish  Color = 220
	ColorOrange    Color = 208
	ColorGray      Color = 250
	ColorWhite     Color = 15
	ColorBlue      Color = 105
	ColorRed       Color = 196
	ColorCoolBlue  Color = 99
	ColorAqua      Color = 122
	ColorDarkOlive Color = 155
	ColorLighSlate Color = 75
	ColorYellow    Color = 226
	ColorYellow2   Color = 190
	ColorGreenPale Color = 114
	ColorGreen     Color = 82
)

// Color tracks the output color.
type Color int

// Colorize a string based on given color.
func Colorize(s string, c Color) string {
	return "\033[38;5;" + strconv.Itoa(int(c)) + ";m" + s + "\033[0m"
}

func colorForLevel(l linter.Level) Color {
	switch l {
	case linter.ErrorLevel:
		return ColorRed
	case linter.WarnLevel:
		return ColorOrangish
	case linter.InfoLevel:
		return ColorAqua
	case linter.OkLevel:
		return ColorDarkOlive
	default:
		return ColorLighSlate
	}
}

func colorForScore(score int) Color {
	switch {
	case score >= 90:
		return ColorGreen
	case score >= 80:
		return ColorGreenPale
	case score >= 70:
		return ColorAqua
	case score >= 60:
		return ColorYellow
	case score >= 50:
		return ColorOrangish
	default:
		return ColorRed
	}
}
