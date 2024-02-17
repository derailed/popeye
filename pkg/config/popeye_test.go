// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import (
	"testing"

	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/types"
	"github.com/stretchr/testify/assert"
)

func TestNewPopeye(t *testing.T) {
	p := NewPopeye()

	assert.Equal(t, 5, p.Resources.Pod.Restarts)

	ok := p.Match(rules.Spec{
		GVR:  types.NewGVR("v1/nodes"),
		FQN:  "-/n1",
		Code: 600,
	})
	assert.False(t, ok)

	ok = p.Match(rules.Spec{
		GVR:  types.NewGVR("v1/namespaces"),
		FQN:  "kube-public",
		Code: 100,
	})
	assert.False(t, ok)
}
