package linter

import (
	"context"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

// PersistentVolumeClaim represents a PersistentVolumeClaim linter.
type PersistentVolumeClaim struct {
	*Linter
}

// NewPersistentVolumeClaim returns a new PersistentVolumeClaim linter.
func NewPersistentVolumeClaim(l Loader, log *zerolog.Logger) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{NewLinter(l, log)}
}

// Lint a PersistentVolumeClaim.
func (p *PersistentVolumeClaim) Lint(ctx context.Context) error {
	pvcs, err := p.ListPersistentVolumeClaims()
	if err != nil {
		return err
	}

	pods, err := p.ListPods()
	if err != nil {
		return nil
	}

	p.lint(pvcs, pods)

	return nil
}

func (p *PersistentVolumeClaim) lint(pvcs map[string]v1.PersistentVolumeClaim, pods map[string]v1.Pod) {
	refs := map[string]struct{}{}
	for fqn, pod := range pods {
		ns, _ := namespaced(fqn)
		for _, v := range pod.Spec.Volumes {
			if v.VolumeSource.PersistentVolumeClaim == nil {
				continue
			}
			refs[ns+"/"+v.VolumeSource.PersistentVolumeClaim.ClaimName] = struct{}{}
		}
	}

	for fqn, pvc := range pvcs {
		p.initIssues(fqn)
		if !p.checkBound(fqn, pvc.Status.Phase) {
			continue
		}
		if _, ok := refs[fqn]; !ok {
			p.addIssue(fqn, WarnLevel, "Used?")
		}
	}
}

func (p *PersistentVolumeClaim) checkBound(fqn string, phase v1.PersistentVolumeClaimPhase) bool {
	switch phase {
	case v1.ClaimPending:
		p.addIssuef(fqn, ErrorLevel, "Pending claim detected")
	case v1.ClaimLost:
		p.addIssuef(fqn, ErrorLevel, "Lost claim detected")
	case v1.ClaimBound:
		return true
	}

	return false
}
