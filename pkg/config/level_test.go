package config_test

import (
	"testing"

	"github.com/derailed/popeye/internal/linter"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestLintLevel(t *testing.T) {
	uu := map[string]linter.Level{
		"ok":    linter.OkLevel,
		"info":  linter.InfoLevel,
		"warn":  linter.WarnLevel,
		"error": linter.ErrorLevel,
		"blee":  linter.OkLevel,
		"":      linter.OkLevel,
	}

	for k, e := range uu {
		assert.Equal(t, e, linter.Level(config.ToLintLevel(&k)))
	}
}
