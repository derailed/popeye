package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
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

// NewPersistentVolumeClaim returns a new sanitizer.
func NewPersistentVolumeClaim(co *issues.Collector, lister PersistentVolumeClaimLister) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{
		Collector:                   co,
		PersistentVolumeClaimLister: lister,
	}
}

// Sanitize cleanse the resource.
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
		ctx = internal.WithFQN(ctx, fqn)
		defer func(fqn string, ctx context.Context) {
			if p.NoConcerns(fqn) && p.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
				p.ClearOutcome(fqn)
			}
		}(fqn, ctx)

		if !p.checkBound(ctx, pvc.Status.Phase) {
			continue
		}
		if _, ok := refs[fqn]; !ok {
			p.AddCode(ctx, 400)
		}
	}

	return nil
}

func (p *PersistentVolumeClaim) checkBound(ctx context.Context, phase v1.PersistentVolumeClaimPhase) bool {
	switch phase {
	case v1.ClaimPending:
		p.AddCode(ctx, 1003)
	case v1.ClaimLost:
		p.AddCode(ctx, 1004)
	case v1.ClaimBound:
		return true
	}

	return false
}
