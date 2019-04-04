package report

import "strings"

// Grade returns a run report grade based on score.
func Grade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	case score >= 50:
		return "E"
	default:
		return "F"
	}
}

// Badge returns a popeye grade.
func Badge(score int) string {
	b := strings.Join(grader, "\n")

	if score < 70 {
		b = strings.Replace(b, "a", "O", 1)
		b = strings.Replace(b, "o", "X", 3)
	}

	return Colorize(strings.Replace(b, "K", Grade(score), 1), colorForScore(score))
}

var grader = []string{
	"o          .-'-.     ",
	" o     __| K    `\\  ",
	"  o   `-,-`--._   `\\",
	" []  .->'  a     `|-'",
	"  `=/ (__/_       /  ",
	"    \\_,    `    _)  ",
	"       `----;  |     ",
}
