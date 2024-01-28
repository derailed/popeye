// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"fmt"
	"strconv"

	"github.com/derailed/popeye/internal/rules"
)

// Color ANSI palette (256!)
const (
	ColorWhite     Color = 15
	ColorLighSlate Color = 75
	ColorGreen     Color = 82
	ColorCoolBlue  Color = 99
	ColorBlue      Color = 105
	ColorGreenPale Color = 114
	ColorAqua      Color = 122
	ColorDarkOlive Color = 155
	ColorYellow2   Color = 190
	ColorRed       Color = 196
	ColorOrangRed  Color = 202
	ColorOrange    Color = 208
	ColorOrangish  Color = 220
	ColorYellow    Color = 226
	ColorGray      Color = 250
)

// Color tracks the output color.
type Color int

// Colorizef colorizes a formatted string.
func Colorizef(c Color, fmat string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(fmat, args...), c)
}

// Colorize a string based on given color.
func Colorize(s string, c Color) string {
	return "\033[38;5;" + strconv.Itoa(int(c)) + "m" + s + "\033[0m"
}

func colorForLevel(l rules.Level) Color {
	switch l {
	case rules.ErrorLevel:
		return ColorRed
	case rules.WarnLevel:
		return ColorOrangish
	case rules.InfoLevel:
		return ColorAqua
	case rules.OkLevel:
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
