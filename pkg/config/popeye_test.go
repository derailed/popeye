package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPopeye(t *testing.T) {
	p := NewPopeye()

	assert.False(t, p.ShouldExclude("node", "n1", 600))
	assert.Equal(t, 5, p.Pod.Restarts)
	assert.False(t, p.ShouldExclude("namespace", "kube-public", 100))
}
