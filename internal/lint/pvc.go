// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

type (
	// PersistentVolumeClaim represents a PersistentVolumeClaim linter.
	PersistentVolumeClaim struct {
		*issues.Collector
		db *db.DB
	}
)

// NewPersistentVolumeClaim returns a new instance.
func NewPersistentVolumeClaim(co *issues.Collector, db *db.DB) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *PersistentVolumeClaim) Lint(ctx context.Context) error {
	refs := make(map[string]struct{})
	txn, it := s.db.MustITFor(internal.Glossary[internal.PO])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		pod := o.(*v1.Pod)
		for _, v := range pod.Spec.Volumes {
			if v.VolumeSource.PersistentVolumeClaim == nil {
				continue
			}
			refs[cache.FQN(pod.Namespace, v.VolumeSource.PersistentVolumeClaim.ClaimName)] = struct{}{}
		}
	}

	txn, it = s.db.MustITFor(internal.Glossary[internal.PVC])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		pvc := o.(*v1.PersistentVolumeClaim)
		fqn := client.FQN(pvc.Namespace, pvc.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, pvc))

		s.checkBound(ctx, pvc.Status.Phase)
		if _, ok := refs[fqn]; !ok {
			s.AddCode(ctx, 400)
		}
	}

	return nil
}

func (s *PersistentVolumeClaim) checkBound(ctx context.Context, phase v1.PersistentVolumeClaimPhase) {
	switch phase {
	case v1.ClaimPending:
		s.AddCode(ctx, 1003)
	case v1.ClaimLost:
		s.AddCode(ctx, 1004)
	}
}
