package report

import (
	"testing"
	"unicode/utf8"

	"github.com/derailed/popeye/internal/linter"
	"github.com/stretchr/testify/assert"
)

func TestEmojiForLevel(t *testing.T) {
	s := new(Sanitizer)
	for k, v := range map[int]int{0: 1, 1: 1, 2: 1, 3: 1, 4: 1} {
		assert.Equal(t, v, utf8.RuneCountInString(s.EmojiForLevel(linter.Level(k))))
	}
}
