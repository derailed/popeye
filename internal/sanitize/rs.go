package sanitize

import (
	"context"
	"errors"

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

		r.checkDeprecation(ctx, rs)

		if r.Config.ExcludeFQN(internal.MustExtractSection(ctx), fqn) {
			r.ClearOutcome(fqn)
		}
	}

	return nil
}

func (r *ReplicaSet) checkDeprecation(ctx context.Context, rs *appsv1.ReplicaSet) {
	const current = "apps/v1"

	rev, err := resourceRev(internal.MustExtractFQN(ctx), rs.Annotations)
	if err != nil {
		rev = revFromLink(rs.SelfLink)
		if rev == "" {
			r.AddCode(ctx, 404, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		r.AddCode(ctx, 403, "ReplicaSet", rev, current)
	}
}
