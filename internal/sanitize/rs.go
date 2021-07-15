package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	appsv1 "k8s.io/api/apps/v1"
)

type (
	// ReplicaSet tracks ReplicaSet sanitization.
	ReplicaSet struct {
		*issues.Collector
		ReplicaSetLister
	}

	// ReplicaLister list replicaset.
	ReplicaLister interface {
		ListReplicaSets() map[string]*appsv1.ReplicaSet
	}

	// ReplicaSetLister list available ReplicaSets on a cluster.
	ReplicaSetLister interface {
		ReplicaLister
	}
)

// NewReplicaSet returns a new ReplicaSet sanitizer.
func NewReplicaSet(co *issues.Collector, lister ReplicaSetLister) *ReplicaSet {
	return &ReplicaSet{
		Collector:        co,
		ReplicaSetLister: lister,
	}
}

// Sanitize cleanse the resource.
func (r *ReplicaSet) Sanitize(ctx context.Context) error {
	for fqn, rs := range r.ListReplicaSets() {
		r.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		r.checkHealth(ctx, rs)
		r.checkDeprecation(ctx, rs)

		if r.NoConcerns(fqn) && r.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			r.ClearOutcome(fqn)
		}
	}

	return nil
}

func (r *ReplicaSet) checkHealth(ctx context.Context, rs *appsv1.ReplicaSet) {
	if rs.Spec.Replicas != nil && *rs.Spec.Replicas != rs.Status.ReadyReplicas {
		r.AddCode(ctx, 1120, *rs.Spec.Replicas, rs.Status.ReadyReplicas)
	}
}

func (r *ReplicaSet) checkDeprecation(ctx context.Context, rs *appsv1.ReplicaSet) {
	const current = "apps/v1"

	rev, err := resourceRev(internal.MustExtractFQN(ctx), "ReplicaSet", rs.Annotations)
	if err != nil {
		if rev = revFromLink(rs.SelfLink); rev == "" {
			return
		}
	}
	if rev != current {
		r.AddCode(ctx, 403, "ReplicaSet", rev, current)
	}
}
