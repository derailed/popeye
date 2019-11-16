package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPVCSanitize(t *testing.T) {
	uu := map[string]struct {
		lister PersistentVolumeClaimLister
		issues int
	}{
		"bound":   {makePVCLister(pvcOpts{used: "pvc1", phase: v1.ClaimBound}), 0},
		"lost":    {makePVCLister(pvcOpts{used: "pvc1", phase: v1.ClaimLost}), 1},
		"pending": {makePVCLister(pvcOpts{used: "pvc1", phase: v1.ClaimPending}), 1},
		"used":    {makePVCLister(pvcOpts{used: "pvc2", phase: v1.ClaimBound}), 1},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			p := NewPersistentVolumeClaim(issues.NewCollector(loadCodes(t)), u.lister)

			assert.Nil(t, p.Sanitize(context.TODO()))
			assert.Equal(t, u.issues, len(p.Outcome()["default/pvc1"]))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type pvcOpts struct {
	phase v1.PersistentVolumeClaimPhase
	used  string
}

type pvc struct {
	name string
	opts pvcOpts
}

func makePVCLister(opts pvcOpts) pvc {
	return pvc{name: "pvc1", opts: opts}
}

func (p pvc) ListPersistentVolumeClaims() map[string]*v1.PersistentVolumeClaim {
	return map[string]*v1.PersistentVolumeClaim{
		"default/pvc1": makePVC(p.opts.used, p.opts.phase),
	}
}

func (p pvc) ListPods() map[string]*v1.Pod {
	return map[string]*v1.Pod{
		"default/p1": makePodPVC("p1", p.opts.used),
	}
}

func (p pvc) GetPod(map[string]string) *v1.Pod {
	return nil
}

func makePVC(n string, p v1.PersistentVolumeClaimPhase) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: p,
		},
	}
}

func makePodPVC(n, pvc string) *v1.Pod {
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
