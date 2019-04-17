package report

import "github.com/derailed/popeye/internal/linter"

const (
	containerLevel linter.Level = 100
	noLevel        linter.Level = 101
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
func (s *Sanitizer) EmojiForLevel(l linter.Level) string {
	var key string

	switch l {
	case noLevel:
		return ""
	case containerLevel:
		key = "container"
	case linter.ErrorLevel:
		key = "farfromfok"
	case linter.WarnLevel:
		key = "warn"
	case linter.InfoLevel:
		key = "fyi"
	default:
		key = "peachy"
	}

	if s.jurassicMode {
		return emojisUgry[key]
	}

	return emojis[key]
}
