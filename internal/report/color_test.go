package report

import (
	"testing"

	"github.com/derailed/popeye/internal/linter"
	"github.com/stretchr/testify/assert"
)

func TestColorForLevel(t *testing.T) {
	for k, v := range map[int]Color{0: ColorDarkOlive, 1: ColorAqua, 2: ColorOrangish, 3: ColorRed} {
		assert.Equal(t, v, colorForLevel(linter.Level(k)))
	}
}
