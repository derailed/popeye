package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	cfg, err := NewConfig(NewFlags())

	assert.Nil(t, err)
	assert.Equal(t, 80.0, cfg.NodeCPULimit())
	assert.Equal(t, 80.0, cfg.NodeMEMLimit())
	assert.Equal(t, 80.0, cfg.PodCPULimit())
	assert.Equal(t, 80.0, cfg.PodMEMLimit())
	assert.False(t, cfg.ShouldExclude("node", "n1"))
	assert.False(t, cfg.ShouldExclude("namespace", "kube-public"))
	assert.False(t, cfg.ShouldExclude("service", "default/kubernetes"))
	assert.Equal(t, 5, cfg.RestartsLimit())
	assert.Equal(t, Allocations{Over: 100, Under: 50}, cfg.CPUResourceLimits())
	assert.Equal(t, Allocations{Over: 100, Under: 50}, cfg.MEMResourceLimits())
	assert.Equal(t, 0, cfg.LinterLevel())
}

func TestNewConfigWithFile(t *testing.T) {
	var (
		dir = "assets/sp1.yml"
		ss  = []string{"s1", "s2"}
		f   = NewFlags()
	)
	f.Sections = &ss
	f.AllNamespaces = boolPtr(true)
	f.Spinach = &dir

	cfg, err := NewConfig(f)
	assert.Nil(t, err)

	assert.Equal(t, 3, cfg.RestartsLimit())
	assert.True(t, cfg.ShouldExclude("node", "n1"))
	assert.False(t, cfg.ShouldExclude("pod", "default/fred"))
	assert.True(t, cfg.ShouldExclude("service", "default/dictionary"))
	assert.True(t, cfg.ShouldExclude("namespace", "kube-public"))
	assert.Equal(t, 90.0, cfg.NodeCPULimit())
	assert.Equal(t, 80.0, cfg.NodeMEMLimit())
	assert.Equal(t, 80.0, cfg.PodCPULimit())
	assert.Equal(t, 75.0, cfg.PodMEMLimit())
	assert.Equal(t, 0, cfg.LintLevel)
	assert.Equal(t, ss, cfg.Sections())
	f.Sections = nil
	assert.Equal(t, []string{}, cfg.Sections())
}

func TestNewConfigNoResourceSpecs(t *testing.T) {
	var (
		dir = "assets/sp2.yml"
		f   = NewFlags()
	)
	f.Spinach = &dir

	cfg, err := NewConfig(f)
	assert.Nil(t, err)

	assert.Equal(t, 80.0, cfg.NodeCPULimit())
	assert.Equal(t, 80.0, cfg.NodeMEMLimit())
	assert.Equal(t, 80.0, cfg.PodCPULimit())
	assert.Equal(t, 80.0, cfg.PodMEMLimit())
}

func TestNewConfigWithFileToast(t *testing.T) {
	var (
		dir = "assets/spinach.yml"
		f   = NewFlags()
	)
	f.Spinach = &dir

	_, err := NewConfig(f)
	assert.NotNil(t, err)
}
