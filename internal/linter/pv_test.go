package linter

import (
	"context"
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPVLinter(t *testing.T) {
	mkl := NewMockLoader()
	m.When(mkl.ListPersistentVolumes()).ThenReturn(map[string]v1.PersistentVolume{
		"pv1": makePV("pv1", v1.VolumeBound),
		"pv2": makePV("pv2", v1.VolumeAvailable),
	}, nil)

	pv := NewPersistentVolume(mkl, nil)
	pv.Lint(context.Background())

	assert.Equal(t, 2, len(pv.Issues()))
	mkl.VerifyWasCalledOnce().ListPersistentVolumes()
}

func TestPVLint(t *testing.T) {
	uu := []struct {
		pvs    map[string]v1.PersistentVolume
		issues int
	}{
		{map[string]v1.PersistentVolume{
			"pv1": makePV("pv1", v1.VolumeBound),
			"pv2": makePV("pv2", v1.VolumeAvailable),
		},
			2,
		},
	}

	for _, u := range uu {
		pv := NewPersistentVolume(nil, nil)
		pv.lint(u.pvs)
		assert.Equal(t, 2, len(pv.Issues()))
	}
}

func TestPVCheckBounds(t *testing.T) {
	uu := []struct {
		phase  v1.PersistentVolumePhase
		level  Level
		issues int
	}{
		{v1.VolumeBound, InfoLevel, 0},
		{v1.VolumeAvailable, InfoLevel, 1},
		{v1.VolumePending, ErrorLevel, 1},
		{v1.VolumeFailed, ErrorLevel, 1},
	}

	fqn := "default/pv1"
	for _, u := range uu {
		p := NewPersistentVolume(nil, nil)
		p.checkBound(fqn, u.phase)

		assert.Equal(t, u.issues, len(p.Issues()[fqn]))
		if u.issues == 1 {
			assert.Equal(t, u.level, p.Issues()[fqn][0].Severity())
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makePV(n string, p v1.PersistentVolumePhase) v1.PersistentVolume {
	return v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
		Status: v1.PersistentVolumeStatus{
			Phase: p,
		},
	}
}
