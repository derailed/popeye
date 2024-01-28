// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

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
func (s *ScanReport) Badge(score int) []string {
	ic := make([]string, len(GraderLogo))
	for i, l := range GraderLogo {
		switch i {
		case 0, 2:
			if score < 70 {
				l = strings.Replace(l, "o", "S", 1)
			}
		case 1:
			l = strings.Replace(l, "K", Grade(score), 1)
		case 3:
			if score < 70 {
				l = strings.Replace(l, "a", "O", 1)
			}
		}
		ic[i] = s.Color(l, colorForScore(score))
	}

	return ic
}

// GraderLogo affords for replacing logo parts.
var GraderLogo = []string{
	"o          .-'-.     ",
	" o     __| K    `\\  ",
	"  o   `-,-`--._   `\\",
	" []  .->'  a     `|-'",
	"  `=/ (__/_       /  ",
	"    \\_,    `    _)  ",
	"       `----;  |     ",
}
