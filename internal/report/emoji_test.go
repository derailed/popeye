package report

import (
	"testing"
	"unicode/utf8"

	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestEmojiForLevel(t *testing.T) {
	for k, v := range map[int]int{0: 1, 1: 1, 2: 1, 3: 1, 4: 1, 100: 1} {
		assert.Equal(t, v, utf8.RuneCountInString(EmojiForLevel(config.Level(k), false)))
	}
}

func TestEmojiUgry(t *testing.T) {
	for k, v := range map[int]string{0: "OK", 1: "I", 2: "W", 3: "E", 100: "C"} {
		assert.Equal(t, v, EmojiForLevel(config.Level(k), true))
	}
}
