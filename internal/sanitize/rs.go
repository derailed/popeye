package sanitize

import (
	"context"
	"errors"

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

// Sanitize configmaps.
func (d *ReplicaSet) Sanitize(ctx context.Context) error {
	for fqn, rs := range d.ListReplicaSets() {
		d.InitOutcome(fqn)
		d.checkDeprecation(fqn, rs)
	}

	return nil
}

func (d *ReplicaSet) checkDeprecation(fqn string, rs *appsv1.ReplicaSet) {
	const current = "apps/v1"

	rev, err := resourceRev(fqn, rs.Annotations)
	if err != nil {
		rev = revFromLink(rs.SelfLink)
		if rev == "" {
			d.AddCode(404, fqn, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		d.AddCode(403, fqn, "ReplicaSet", rev, current)
	}
}

// CheckReplicaSet checks if deployment contract is currently happy or not.
func (d *ReplicaSet) checkReplicaSet(fqn string, rs *appsv1.ReplicaSet) {
	if rs.Spec.Replicas == nil || (rs.Spec.Replicas != nil && *rs.Spec.Replicas == 0) {
		d.AddCode(500, fqn)
	}

	if rs.Status.AvailableReplicas == 0 {
		d.AddCode(501, fqn)
	}

}

// Helpers...
