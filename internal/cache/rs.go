package cache

import (
	appsv1 "k8s.io/api/apps/v1"
)

// ReplicaSetKey tracks ReplicaSet resource references
const ReplicaSetKey = "ds"

// ReplicaSet represents ReplicaSet cache.
type ReplicaSet struct {
	rss map[string]*appsv1.ReplicaSet
}

// NewReplicaSet returns a new ReplicaSet cache.
func NewReplicaSet(rss map[string]*appsv1.ReplicaSet) *ReplicaSet {
	return &ReplicaSet{rss: rss}
}

// ListReplicaSets returns all available ReplicaSets on the cluster.
func (d *ReplicaSet) ListReplicaSets() map[string]*appsv1.ReplicaSet {
	return d.rss
}
