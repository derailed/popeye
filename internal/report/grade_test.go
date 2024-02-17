// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBadge(t *testing.T) {
	uu := []struct {
		score int
		e     string
	}{
		{
			90,
			"\x1b[38;5;82mo          .-'-.     \x1b[0m\n\x1b[38;5;82m o     __| A    `\\  \x1b[0m\n\x1b[38;5;82m  o   `-,-`--._   `\\\x1b[0m\n\x1b[38;5;82m []  .->'  a     `|-'\x1b[0m\n\x1b[38;5;82m  `=/ (__/_       /  \x1b[0m\n\x1b[38;5;82m    \\_,    `    _)  \x1b[0m\n\x1b[38;5;82m       `----;  |     \x1b[0m",
		},
		{
			80,
			"\x1b[38;5;114mo          .-'-.     \x1b[0m\n\x1b[38;5;114m o     __| B    `\\  \x1b[0m\n\x1b[38;5;114m  o   `-,-`--._   `\\\x1b[0m\n\x1b[38;5;114m []  .->'  a     `|-'\x1b[0m\n\x1b[38;5;114m  `=/ (__/_       /  \x1b[0m\n\x1b[38;5;114m    \\_,    `    _)  \x1b[0m\n\x1b[38;5;114m       `----;  |     \x1b[0m",
		},
		{
			70,
			"\x1b[38;5;122mo          .-'-.     \x1b[0m\n\x1b[38;5;122m o     __| C    `\\  \x1b[0m\n\x1b[38;5;122m  o   `-,-`--._   `\\\x1b[0m\n\x1b[38;5;122m []  .->'  a     `|-'\x1b[0m\n\x1b[38;5;122m  `=/ (__/_       /  \x1b[0m\n\x1b[38;5;122m    \\_,    `    _)  \x1b[0m\n\x1b[38;5;122m       `----;  |     \x1b[0m",
		},
		{
			60,
			"\x1b[38;5;226mS          .-'-.     \x1b[0m\n\x1b[38;5;226m o     __| D    `\\  \x1b[0m\n\x1b[38;5;226m  S   `-,-`--._   `\\\x1b[0m\n\x1b[38;5;226m []  .->'  O     `|-'\x1b[0m\n\x1b[38;5;226m  `=/ (__/_       /  \x1b[0m\n\x1b[38;5;226m    \\_,    `    _)  \x1b[0m\n\x1b[38;5;226m       `----;  |     \x1b[0m",
		},
		{
			50,
			"\x1b[38;5;220mS          .-'-.     \x1b[0m\n\x1b[38;5;220m o     __| E    `\\  \x1b[0m\n\x1b[38;5;220m  S   `-,-`--._   `\\\x1b[0m\n\x1b[38;5;220m []  .->'  O     `|-'\x1b[0m\n\x1b[38;5;220m  `=/ (__/_       /  \x1b[0m\n\x1b[38;5;220m    \\_,    `    _)  \x1b[0m\n\x1b[38;5;220m       `----;  |     \x1b[0m",
		},
		{
			40,
			"\x1b[38;5;196mS          .-'-.     \x1b[0m\n\x1b[38;5;196m o     __| F    `\\  \x1b[0m\n\x1b[38;5;196m  S   `-,-`--._   `\\\x1b[0m\n\x1b[38;5;196m []  .->'  O     `|-'\x1b[0m\n\x1b[38;5;196m  `=/ (__/_       /  \x1b[0m\n\x1b[38;5;196m    \\_,    `    _)  \x1b[0m\n\x1b[38;5;196m       `----;  |     \x1b[0m",
		},
	}

	s := new(ScanReport)
	for _, u := range uu {
		assert.Equal(t, u.e, strings.Join(s.Badge(u.score), "\n"))
	}
}
