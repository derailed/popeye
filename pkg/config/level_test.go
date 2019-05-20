package config

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

	for k, e := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, e, ToIssueLevel(&k))
		})
	}
}
