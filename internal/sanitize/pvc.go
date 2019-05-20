package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

type (
	// PersistentVolumeClaimLister list available PersistentVolumeClaim on a cluster.
	PersistentVolumeClaimLister interface {
		ListPersistentVolumeClaims() map[string]*v1.PersistentVolumeClaim
		PodLister
	}

	// PersistentVolumeClaim represents a PersistentVolumeClaim sanitizer.
	PersistentVolumeClaim struct {
		*issues.Collector
		PersistentVolumeClaimLister
	}
)

// NewPersistentVolumeClaim returns a new PersistentVolumeClaim sanitizer.
func NewPersistentVolumeClaim(co *issues.Collector, lister PersistentVolumeClaimLister) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{
		Collector:                   co,
		PersistentVolumeClaimLister: lister,
	}
}

// Sanitize a PersistentVolumeClaim.
func (p *PersistentVolumeClaim) Sanitize(ctx context.Context) error {
	refs := map[string]struct{}{}
	for fqn, pod := range p.ListPods() {
		ns, _ := namespaced(fqn)
		for _, v := range pod.Spec.Volumes {
			if v.VolumeSource.PersistentVolumeClaim == nil {
				continue
			}
			refs[cache.FQN(ns, v.VolumeSource.PersistentVolumeClaim.ClaimName)] = struct{}{}
		}
	}

	for fqn, pvc := range p.ListPersistentVolumeClaims() {
		p.InitOutcome(fqn)
		if !p.checkBound(fqn, pvc.Status.Phase) {
			continue
		}
		if _, ok := refs[fqn]; !ok {
			p.AddWarn(fqn, "Used?")
		}
	}

	return nil
}

func (p *PersistentVolumeClaim) checkBound(fqn string, phase v1.PersistentVolumeClaimPhase) bool {
	switch phase {
	case v1.ClaimPending:
		p.AddError(fqn, "Pending claim detected")
	case v1.ClaimLost:
		p.AddError(fqn, "Lost claim detected")
	case v1.ClaimBound:
		return true
	}

	return false
}
