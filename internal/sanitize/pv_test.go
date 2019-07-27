package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPVSanitize(t *testing.T) {
	uu := map[string]struct {
		lister PersistentVolumeLister
		issues int
	}{
		"bound":     {makePVLister("pv1", pvOpts{phase: v1.VolumeBound}), 0},
		"available": {makePVLister("pv1", pvOpts{phase: v1.VolumeAvailable}), 1},
		"pending":   {makePVLister("pv1", pvOpts{phase: v1.VolumePending}), 1},
		"failed":    {makePVLister("pv1", pvOpts{phase: v1.VolumeFailed}), 1},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			p := NewPersistentVolume(issues.NewCollector(loadCodes(t)), u.lister)
			p.Sanitize(context.Background())

			assert.Equal(t, u.issues, len(p.Outcome()["default/pv1"]))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type pvOpts struct {
	phase v1.PersistentVolumePhase
}

type pv struct {
	name string
	opts pvOpts
}

func makePVLister(n string, opts pvOpts) pv {
	return pv{name: n, opts: opts}
}

func (p pv) ListPersistentVolumes() map[string]*v1.PersistentVolume {
	return map[string]*v1.PersistentVolume{
		"default/pv1": makePV(p.name, p.opts.phase),
	}
}

func makePV(n string, p v1.PersistentVolumePhase) *v1.PersistentVolume {
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
		Status: v1.PersistentVolumeStatus{
			Phase: p,
		},
	}
}
