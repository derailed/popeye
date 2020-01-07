package cache

import (
	appsv1 "k8s.io/api/apps/v1"
)

// DaemonSetKey tracks DaemonSet resource references
const DaemonSetKey = "ds"

// DaemonSet represents DaemonSet cache.
type DaemonSet struct {
	ds map[string]*appsv1.DaemonSet
}

// NewDaemonSet returns a new DaemonSet cache.
func NewDaemonSet(ds map[string]*appsv1.DaemonSet) *DaemonSet {
	return &DaemonSet{ds: ds}
}

// ListDaemonSets returns all available DaemonSets on the cluster.
func (d *DaemonSet) ListDaemonSets() map[string]*appsv1.DaemonSet {
	return d.ds
}
