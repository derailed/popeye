// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"testing"

	"github.com/derailed/popeye/internal/rules"
	"github.com/stretchr/testify/assert"
)

func TestColorForLevel(t *testing.T) {
	colors := map[int]Color{
		0: ColorDarkOlive,
		1: ColorAqua,
		2: ColorOrangish,
		3: ColorRed,
		4: ColorLighSlate,
	}

	for k, v := range colors {
		assert.Equal(t, v, colorForLevel(rules.Level(k)))
	}
}
