// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLintLevel(t *testing.T) {
	uu := map[string]Level{
		"ok":    OkLevel,
		"info":  InfoLevel,
		"warn":  WarnLevel,
		"error": ErrorLevel,
		"blee":  OkLevel,
		"":      OkLevel,
	}

	for k := range uu {
		u, key := uu[k], k
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u, ToIssueLevel(&key))
		})
	}
}
