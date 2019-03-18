package linter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNsCheckActive(t *testing.T) {
	uu := []struct {
		phase  v1.NamespacePhase
		issues int
	}{
		{v1.NamespaceActive, 0},
		{v1.NamespaceTerminating, 1},
	}

	for _, u := range uu {
		ns := v1.Namespace{
			Status: v1.NamespaceStatus{
				Phase: u.phase,
			},
		}
		l := NewNamespace()
		l.Lint(ns)
		assert.Equal(t, u.issues, len(l.Issues()))
	}
}
