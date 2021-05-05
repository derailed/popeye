package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

type (
	// PersistentVolumeLister list available PersistentVolume on a cluster.
	PersistentVolumeLister interface {
		ListPersistentVolumes() map[string]*v1.PersistentVolume
	}

	// PersistentVolume represents a PersistentVolume sanitizer.
	PersistentVolume struct {
		*issues.Collector
		PersistentVolumeLister
	}
)

// NewPersistentVolume returns a new sanitizer.
func NewPersistentVolume(co *issues.Collector, lister PersistentVolumeLister) *PersistentVolume {
	return &PersistentVolume{
		Collector:              co,
		PersistentVolumeLister: lister,
	}
}

// Sanitize cleanse the resource.
func (p *PersistentVolume) Sanitize(ctx context.Context) error {
	for fqn, pv := range p.ListPersistentVolumes() {
		p.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		p.checkBound(ctx, pv.Status.Phase)

		if p.NoConcerns(fqn) && p.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			p.ClearOutcome(fqn)
		}
	}

	return nil
}

func (p *PersistentVolume) checkBound(ctx context.Context, phase v1.PersistentVolumePhase) {
	// nolint:exhaustive
	switch phase {
	case v1.VolumeAvailable:
		p.AddCode(ctx, 1000)
	case v1.VolumePending:
		p.AddCode(ctx, 1001)
	case v1.VolumeFailed:
		p.AddCode(ctx, 1002)
	}
}
