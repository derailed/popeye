package report

import "github.com/derailed/popeye/internal/linter"

const containerLevel linter.Level = 100

var emojis = map[string]string{
	"peachy":     "✅",
	"farfromfok": "💥",
	"warn":       "😱",
	"fyi":        "🔊",
	"container":  "🐳",
}

// EmojiForLevel maps lint levels to emojis.
func EmojiForLevel(l linter.Level) string {
	switch l {
	case containerLevel:
		return emojis["container"]
	case linter.ErrorLevel:
		return emojis["farfromfok"]
	case linter.WarnLevel:
		return emojis["warn"]
	case linter.InfoLevel:
		return emojis["fyi"]
	default:
		return emojis["peachy"]
	}
}
