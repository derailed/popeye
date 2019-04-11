package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPopeye(t *testing.T) {
	p := NewPopeye()

	assert.False(t, p.Node.excluded("n1"))
	assert.Equal(t, 5, p.Pod.Restarts)
	assert.True(t, p.Namespace.excluded("kube-public"))
}
