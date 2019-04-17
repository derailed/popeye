package config

import (
	"testing"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	cfg, err := NewConfig(k8s.NewFlags())

	assert.Nil(t, err)
	assert.Equal(t, 80.0, cfg.NodeCPULimit())
	assert.Equal(t, 80.0, cfg.NodeMEMLimit())
	assert.Equal(t, 80.0, cfg.PodCPULimit())
	assert.Equal(t, 80.0, cfg.PodMEMLimit())
	assert.False(t, cfg.ExcludedNode("n1"))
	assert.True(t, cfg.ExcludedNS("kube-public"))
	assert.True(t, cfg.ExcludedService("default/kubernetes"))
	assert.Equal(t, 5, cfg.RestartsLimit())
	assert.Equal(t, Allocations{Over: 100, Under: 50}, cfg.CPUResourceLimits())
	assert.Equal(t, Allocations{Over: 100, Under: 50}, cfg.MEMResourceLimits())
	assert.Equal(t, 0, cfg.LinterLevel())
}

func TestNewConfigWithFile(t *testing.T) {
	dir := "assets/sp1.yml"

	f := k8s.NewFlags()
	f.Spinach = &dir

	cfg, err := NewConfig(f)

	assert.Nil(t, err)
	assert.Equal(t, 3, cfg.RestartsLimit())
	assert.True(t, cfg.ExcludedNode("n1"))
	assert.False(t, cfg.ExcludedService("default/fred"))
	assert.True(t, cfg.ExcludedService("default/dictionary"))
	assert.True(t, cfg.ExcludedNS("kube-public"))
	assert.Equal(t, 90.0, cfg.NodeCPULimit())
	assert.Equal(t, 80.0, cfg.NodeMEMLimit())
	assert.Equal(t, 80.0, cfg.PodCPULimit())
	assert.Equal(t, 75.0, cfg.PodMEMLimit())
	assert.Equal(t, 0, cfg.LintLevel)
	assert.Equal(t, []string{}, cfg.Sections())
}

func TestNewConfigWithFileToast(t *testing.T) {
	dir := "assets/spinach.yml"

	f := k8s.NewFlags()
	f.Spinach = &dir

	_, err := NewConfig(f)
	assert.NotNil(t, err)
}
