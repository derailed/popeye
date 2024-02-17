// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config_test

import (
	"testing"

	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestConfigGlobalExcludes(t *testing.T) {
	uu := map[string]struct {
		spec rules.Spec
		e    bool
	}{
		"exact-ns": {
			spec: rules.Spec{
				GVR:        types.NewGVR("v1/pods"),
				FQN:        "gns1/blee",
				Containers: []string{"c1"},
			},
			e: true,
		},
	}

	sp := "testdata/sp3.yml"
	f := config.NewFlags()
	f.Spinach = &sp
	cfg, err := config.NewConfig(f)
	assert.NoError(t, err)

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, cfg.Match(u.spec))
		})
	}
}

func TestConfigExcludes(t *testing.T) {
	uu := map[string]struct {
		spec rules.Spec
		e    bool
	}{
		"exact-ns": {
			spec: rules.Spec{
				GVR:        types.NewGVR("v1/pods"),
				FQN:        "ns1/blee",
				Containers: []string{"c1"},
			},
			e: true,
		},
		"skip-exact-ns": {
			spec: rules.Spec{
				GVR: types.NewGVR("v1/pods"),
				FQN: "ns5/blee",
			},
		},

		"skip-gvr-no-rules": {
			spec: rules.Spec{
				GVR: types.NewGVR("v1/configmaps"),
				FQN: "fred/cm1",
			},
		},

		"match-annotations": {
			spec: rules.Spec{
				GVR:         types.NewGVR("v1/nodes"),
				Annotations: rules.Labels{"a1": "b1"},
				Code:        100,
			},
			e: true,
		},
		"skip-annotations-key": {
			spec: rules.Spec{
				GVR:         types.NewGVR("v1/nodes"),
				FQN:         "fred/annot1",
				Annotations: rules.Labels{"a5": "b1"},
				Code:        100,
			},
		},
		"skip-annotations-val": {
			spec: rules.Spec{
				GVR:         types.NewGVR("v1/nodes"),
				FQN:         "fred/annot2",
				Annotations: rules.Labels{"a1": "b2"},
				Code:        100,
			},
		},

		"match-labels": {
			spec: rules.Spec{
				GVR:    types.NewGVR("v1/pods"),
				FQN:    "fred/bozo",
				Labels: rules.Labels{"kube-system": "fred"},
				Code:   300,
			},
			e: true,
		},
		"skip-labels": {
			spec: rules.Spec{
				GVR:    types.NewGVR("v1/pods"),
				FQN:    "fred/bozo",
				Labels: rules.Labels{"kube-system": "fred1"},
				Code:   300,
			},
		},

		"exact-container": {
			spec: rules.Spec{
				GVR:        types.NewGVR("v1/pods"),
				FQN:        "ns1/blee",
				Containers: []string{"c1"},
			},
			e: true,
		},
		"skip-container": {
			spec: rules.Spec{
				GVR:        types.NewGVR("v1/pods"),
				FQN:        "ns1/blee-1",
				Containers: []string{"bozo"},
			},
		},

		"exact-code": {
			spec: rules.Spec{
				GVR:    types.NewGVR("v1/services"),
				FQN:    "blee/svc1",
				Labels: rules.Labels{"default": "dictionary"},
				Code:   100,
			},
			e: true,
		},
		"skip-exact-code": {
			spec: rules.Spec{
				GVR:  types.NewGVR("v1/services"),
				FQN:  "blee/svc2",
				Code: 301,
			},
		},

		"regex-start": {
			spec: rules.Spec{
				GVR: types.NewGVR("v1/pods"),
				FQN: "istio/fred",
			},
			e: true,
		},
		"skip-regex-start": {
			spec: rules.Spec{
				GVR: types.NewGVR("v1/pods"),
				FQN: "ns-10/fred",
			},
		},
		"regex-contains": {
			spec: rules.Spec{
				GVR: types.NewGVR("v1/secrets"),
				FQN: "istio-fred",
			},
			e: true,
		},
		"skip-regex-contains": {
			spec: rules.Spec{
				GVR: types.NewGVR("v1/secrets"),
				FQN: "click-clack/bozo",
			},
		},
	}

	sp := "testdata/sp3.yml"
	f := config.NewFlags()
	f.Spinach = &sp
	cfg, err := config.NewConfig(f)
	assert.NoError(t, err)

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, cfg.Match(u.spec))
		})
	}
}

func TestNewConfig(t *testing.T) {
	cfg, err := config.NewConfig(config.NewFlags())

	assert.Nil(t, err)
	assert.Equal(t, 80.0, cfg.NodeCPULimit())
	assert.Equal(t, 80.0, cfg.NodeMEMLimit())
	assert.Equal(t, 80.0, cfg.PodCPULimit())
	assert.Equal(t, 80.0, cfg.PodMEMLimit())

	ok := cfg.Match(rules.Spec{
		GVR:  types.NewGVR("v1/nodes"),
		FQN:  "no1",
		Code: 100,
	})
	assert.False(t, ok)

	ok = cfg.Match(rules.Spec{
		GVR:  types.NewGVR("v1/namespaces"),
		FQN:  "kube-public",
		Code: 100,
	})
	assert.False(t, ok)

	ok = cfg.Match(rules.Spec{
		GVR:  types.NewGVR("v1/services"),
		FQN:  "default/svc1",
		Code: 100,
	})
	assert.False(t, ok)

	assert.Equal(t, 5, cfg.RestartsLimit())
	assert.Equal(t, config.Allocations{UnderPerc: 200, OverPerc: 50}, cfg.CPUResourceLimits())
	assert.Equal(t, config.Allocations{UnderPerc: 200, OverPerc: 50}, cfg.MEMResourceLimits())
	assert.Equal(t, 0, cfg.LintLevel)
	assert.Nil(t, cfg.Registries)
}

func TestNewConfigWithFile(t *testing.T) {
	var (
		dir  = "testdata/sp1.yml"
		ss   = []string{"s1", "s2"}
		f    = config.NewFlags()
		true = true
	)
	f.Sections = &ss
	f.AllNamespaces = &true
	f.Spinach = &dir

	cfg, err := config.NewConfig(f)
	assert.Nil(t, err)

	assert.Equal(t, 3, cfg.RestartsLimit())

	ok := cfg.Match(rules.Spec{
		GVR:    types.NewGVR("v1/nodes"),
		FQN:    "n1",
		Code:   100,
		Labels: rules.Labels{"fred": "blee", "k8s-app": "app1"},
	})
	assert.True(t, ok)

	ok = cfg.Match(rules.Spec{
		GVR:  types.NewGVR("v1/pods"),
		FQN:  "default/fred",
		Code: 111,
	})
	assert.False(t, ok)

	ok = cfg.Match(rules.Spec{
		GVR:  types.NewGVR("v1/services"),
		FQN:  "default/dictionary",
		Code: 100,
	})
	assert.False(t, ok)

	ok = cfg.Match(rules.Spec{
		GVR:  types.NewGVR("v1/namespaces"),
		FQN:  "kube-public",
		Code: 100,
	})
	assert.False(t, ok)

	assert.Equal(t, 90.0, cfg.NodeCPULimit())
	assert.Equal(t, 80.0, cfg.NodeMEMLimit())
	assert.Equal(t, 80.0, cfg.PodCPULimit())
	assert.Equal(t, 75.0, cfg.PodMEMLimit())
	assert.Equal(t, 0, cfg.LintLevel)
	assert.Equal(t, ss, cfg.Sections())
	f.Sections = nil
	assert.Equal(t, []string{}, cfg.Sections())
	assert.Equal(t, []string{"docker.io"}, cfg.Registries)
}

func TestNewConfigNoResourceSpec(t *testing.T) {
	var (
		dir = "testdata/sp2.yml"
		f   = config.NewFlags()
	)
	f.Spinach = &dir

	cfg, err := config.NewConfig(f)
	assert.Nil(t, err)

	assert.Equal(t, 80.0, cfg.NodeCPULimit())
	assert.Equal(t, 80.0, cfg.NodeMEMLimit())
	assert.Equal(t, 80.0, cfg.PodCPULimit())
	assert.Equal(t, 80.0, cfg.PodMEMLimit())
}

func TestNewConfigFileToast(t *testing.T) {
	var (
		dir = "testdata/sp-toast.yml"
		f   = config.NewFlags()
	)
	f.Spinach = &dir

	_, err := config.NewConfig(f)
	assert.NotNil(t, err)
}

func TestNewConfigFileNoExists(t *testing.T) {
	var (
		dir = "testdata/spinach.yml"
		f   = config.NewFlags()
	)
	f.Spinach = &dir

	_, err := config.NewConfig(f)
	assert.NotNil(t, err)
}
