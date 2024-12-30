// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	appsv1 "k8s.io/api/apps/v1"
)

// ReplicaSet tracks ReplicaSet sanitization.
type ReplicaSet struct {
	*issues.Collector

	db *db.DB
}

// NewReplicaSet returns a new instance.
func NewReplicaSet(co *issues.Collector, db *db.DB) *ReplicaSet {
	return &ReplicaSet{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *ReplicaSet) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[internal.RS])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		rs := o.(*appsv1.ReplicaSet)
		fqn := client.FQN(rs.Namespace, rs.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, coSpecFor(fqn, rs, rs.Spec.Template.Spec))

		s.checkHealth(ctx, rs)
	}

	return nil
}

func (s *ReplicaSet) checkHealth(ctx context.Context, rs *appsv1.ReplicaSet) {
	if rs.Spec.Replicas != nil && *rs.Spec.Replicas != rs.Status.ReadyReplicas {
		s.AddCode(ctx, 1120, *rs.Spec.Replicas, rs.Status.ReadyReplicas)
	}
}
