// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

type (
	// PersistentVolume represents a PersistentVolume linter.
	PersistentVolume struct {
		*issues.Collector
		db *db.DB
	}
)

// NewPersistentVolume returns a new instance.
func NewPersistentVolume(co *issues.Collector, db *db.DB) *PersistentVolume {
	return &PersistentVolume{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *PersistentVolume) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[internal.PV])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		pv := o.(*v1.PersistentVolume)
		fqn := client.FQN(pv.Namespace, pv.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, pv))

		s.checkBound(ctx, pv.Status.Phase)
	}

	return nil
}

func (s *PersistentVolume) checkBound(ctx context.Context, phase v1.PersistentVolumePhase) {
	switch phase {
	case v1.VolumeAvailable:
		s.AddCode(ctx, 1000)
	case v1.VolumePending:
		s.AddCode(ctx, 1001)
	case v1.VolumeFailed:
		s.AddCode(ctx, 1002)
	}
}
