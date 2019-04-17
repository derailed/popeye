package linter

import (
	"context"
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPVCLinter(t *testing.T) {
	mkl := NewMockLoader()
	m.When(mkl.ListPersistentVolumeClaims()).ThenReturn(map[string]v1.PersistentVolumeClaim{
		"default/pvc1": makePVC("pvc1", v1.ClaimBound),
		"default/pvc2": makePVC("pvc2", v1.ClaimLost),
	}, nil)
	m.When(mkl.ListPods()).ThenReturn(map[string]v1.Pod{
		"default/p1": makePodPVC("p1", "pvc1"),
		"default/p2": makePodPVC("p2", "pvc2"),
	}, nil)

	pv := NewPersistentVolumeClaim(mkl, nil)
	pv.Lint(context.Background())

	assert.Equal(t, 2, len(pv.Issues()))
	mkl.VerifyWasCalledOnce().ListPersistentVolumeClaims()
	mkl.VerifyWasCalledOnce().ListPods()
}

func TestPVCLint(t *testing.T) {
	uu := []struct {
		pvcs   map[string]v1.PersistentVolumeClaim
		pods   map[string]v1.Pod
		issues int
	}{
		{
			map[string]v1.PersistentVolumeClaim{
				"default/pvc1": makePVC("pvc1", v1.ClaimBound),
				"default/pvc2": makePVC("pvc2", v1.ClaimBound),
				"default/pvc3": makePVC("pvc3", v1.ClaimBound),
			},
			map[string]v1.Pod{
				"default/p1": makePodPVC("p1", "pvc1"),
				"default/p2": makePodPVC("p2", "pvc2"),
				"default/p3": makePodVolume("p3", "cm1", "fred", false),
			},
			1,
		},
	}

	for _, u := range uu {
		p := NewPersistentVolumeClaim(nil, nil)
		p.lint(u.pvcs, u.pods)

		assert.Equal(t, 3, len(p.Issues()))
		assert.Equal(t, 0, len(p.Issues()["default/pvc1"]))
		assert.Equal(t, 0, len(p.Issues()["default/pvc2"]))
		assert.Equal(t, u.issues, len(p.Issues()["default/pvc3"]))
	}
}

func TestPVCCheckBounds(t *testing.T) {
	uu := []struct {
		phase  v1.PersistentVolumeClaimPhase
		level  Level
		issues int
	}{
		{v1.ClaimBound, InfoLevel, 0},
		{v1.ClaimLost, ErrorLevel, 1},
		{v1.ClaimPending, ErrorLevel, 1},
	}

	fqn := "default/pv1"
	for _, u := range uu {
		p := NewPersistentVolumeClaim(nil, nil)
		p.checkBound(fqn, u.phase)

		assert.Equal(t, u.issues, len(p.Issues()[fqn]))
		if u.issues == 1 {
			assert.Equal(t, u.level, p.Issues()[fqn][0].Severity())
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makePVC(n string, p v1.PersistentVolumeClaimPhase) v1.PersistentVolumeClaim {
	return v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: p,
		},
	}
}

func makePodPVC(n, pvc string) v1.Pod {
	po := makePod(n)
	po.Spec.Volumes = []v1.Volume{
		{
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc,
				},
			},
		},
	}

	return po
}
