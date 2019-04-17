package linter

import (
	"context"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

// PersistentVolume represents a PersistentVolume linter.
type PersistentVolume struct {
	*Linter
}

// NewPersistentVolume returns a new PersistentVolume linter.
func NewPersistentVolume(l Loader, log *zerolog.Logger) *PersistentVolume {
	return &PersistentVolume{NewLinter(l, log)}
}

// Lint a PersistentVolume.
func (p *PersistentVolume) Lint(ctx context.Context) error {
	pvs, err := p.ListPersistentVolumes()
	if err != nil {
		return err
	}

	p.lint(pvs)

	return nil
}

func (p *PersistentVolume) lint(pvs map[string]v1.PersistentVolume) {
	for fqn, pv := range pvs {
		p.initIssues(fqn)
		p.checkBound(fqn, pv.Status.Phase)
	}
}

func (p *PersistentVolume) checkBound(fqn string, phase v1.PersistentVolumePhase) {
	switch phase {
	case v1.VolumeAvailable:
		p.addIssuef(fqn, InfoLevel, "Available")
	case v1.VolumePending:
		p.addIssuef(fqn, ErrorLevel, "Pending volume detected")
	case v1.VolumeFailed:
		p.addIssuef(fqn, ErrorLevel, "Lost volume detected")
	}
}
