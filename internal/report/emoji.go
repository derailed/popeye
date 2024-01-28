// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"github.com/derailed/popeye/internal/rules"
)

const (
	containerLevel rules.Level = 100
)

var emojis = map[string]string{
	"peachy":     "✅",
	"farfromfok": "💥",
	"warn":       "😱",
	"fyi":        "🔊",
	"container":  "🐳",
}

var emojisUgry = map[string]string{
	"peachy":     "OK",
	"farfromfok": "E",
	"warn":       "W",
	"fyi":        "I",
	"container":  "C",
}

// EmojiForLevel maps lint levels to emojis.
func EmojiForLevel(l rules.Level, jurassic bool) string {
	var key string
	// nolint:exhaustive
	switch l {
	case containerLevel:
		key = "container"
	case rules.ErrorLevel:
		key = "farfromfok"
	case rules.WarnLevel:
		key = "warn"
	case rules.InfoLevel:
		key = "fyi"
	default:
		key = "peachy"
	}

	if jurassic {
		return emojisUgry[key]
	}

	return emojis[key]
}
