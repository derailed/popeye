package sanitize

import (
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
		"bound":     {makePVLister(pvOpts{phase: v1.VolumeBound}), 0},
		"available": {makePVLister(pvOpts{phase: v1.VolumeAvailable}), 1},
		"pending":   {makePVLister(pvOpts{phase: v1.VolumePending}), 1},
		"failed":    {makePVLister(pvOpts{phase: v1.VolumeFailed}), 1},
	}

	ctx := makeContext("v1/persistentvolumes", "pv")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := NewPersistentVolume(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, p.Sanitize(ctx))
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

func makePVLister(opts pvOpts) pv {
	return pv{name: "pv1", opts: opts}
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
