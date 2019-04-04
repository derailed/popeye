package config_test

import (
	"testing"

	"github.com/rs/zerolog"

	"github.com/derailed/popeye/internal/config"
	"github.com/derailed/popeye/internal/linter"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestInit(t *testing.T) {
	cfg := config.New()

	flags := genericclioptions.ConfigFlags{}

	assert.Nil(t, cfg.Init(&flags))
	assert.Equal(t, 80.0, cfg.NodeCPULimit())
	assert.Equal(t, 80.0, cfg.NodeMEMLimit())
	assert.Equal(t, 80.0, cfg.PodCPULimit())
	assert.Equal(t, 80.0, cfg.PodMEMLimit())
	assert.False(t, cfg.ExcludedNode("n1"))
	assert.False(t, cfg.ExcludedService("default/kubernetes"))
	assert.False(t, cfg.ExcludedNS("kube-public"))
	assert.Equal(t, 5, cfg.RestartsLimit())
	assert.Equal(t, "", cfg.ActiveNamespace())
}

func TestLogLevel(t *testing.T) {
	uu := map[string]zerolog.Level{
		"debug": zerolog.DebugLevel,
		"warn":  zerolog.WarnLevel,
		"error": zerolog.ErrorLevel,
		"fatal": zerolog.FatalLevel,
		"blee":  zerolog.InfoLevel,
		"":      zerolog.InfoLevel,
	}

	for k, e := range uu {
		assert.Equal(t, e, config.ToLogLevel(k))
	}
}

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
		assert.Equal(t, e, linter.Level(config.ToLintLevel(k)))
	}
}
