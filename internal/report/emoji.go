package report

import (
	"github.com/derailed/popeye/internal/issues"
)

const (
	containerLevel issues.Level = 100
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
func EmojiForLevel(l issues.Level, jurassic bool) string {
	var key string
	switch l {
	case containerLevel:
		key = "container"
	case issues.ErrorLevel:
		key = "farfromfok"
	case issues.WarnLevel:
		key = "warn"
	case issues.InfoLevel:
		key = "fyi"
	default:
		key = "peachy"
	}

	if jurassic {
		return emojisUgry[key]
	}

	return emojis[key]
}
