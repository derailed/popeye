package sanitize

import (
	"context"

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

// NewPersistentVolume returns a new PersistentVolume sanitizer.
func NewPersistentVolume(co *issues.Collector, lister PersistentVolumeLister) *PersistentVolume {
	return &PersistentVolume{
		Collector:              co,
		PersistentVolumeLister: lister,
	}
}

// Sanitize a PersistentVolume.
func (p *PersistentVolume) Sanitize(ctx context.Context) error {
	for fqn, pv := range p.ListPersistentVolumes() {
		p.InitOutcome(fqn)
		p.checkBound(fqn, pv.Status.Phase)
	}

	return nil
}

func (p *PersistentVolume) checkBound(fqn string, phase v1.PersistentVolumePhase) {
	switch phase {
	case v1.VolumeAvailable:
		p.AddInfo(fqn, "Available")
	case v1.VolumePending:
		p.AddError(fqn, "Pending volume detected")
	case v1.VolumeFailed:
		p.AddError(fqn, "Lost volume detected")
	}
}
