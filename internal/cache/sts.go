package cache

import (
	appsv1 "k8s.io/api/apps/v1"
)

// StatefulSet represents a collection of StatefulSets available on a cluster.
type StatefulSet struct {
	sts map[string]*appsv1.StatefulSet
}

// NewStatefulSet returns a new StatefulSet.
func NewStatefulSet(sts map[string]*appsv1.StatefulSet) *StatefulSet {
	return &StatefulSet{sts}
}

// ListStatefulSets returns all available StatefulSets on the cluster.
func (s *StatefulSet) ListStatefulSets() map[string]*appsv1.StatefulSet {
	return s.sts
}
