package report

import (
	"github.com/derailed/popeye/pkg/config"
)

const (
	containerLevel config.Level = 100
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
func EmojiForLevel(l config.Level, jurassic bool) string {
	var key string
	// nolint:exhaustive
	switch l {
	case containerLevel:
		key = "container"
	case config.ErrorLevel:
		key = "farfromfok"
	case config.WarnLevel:
		key = "warn"
	case config.InfoLevel:
		key = "fyi"
	default:
		key = "peachy"
	}

	if jurassic {
		return emojisUgry[key]
	}

	return emojis[key]
}
