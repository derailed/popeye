package cache

import (
	v1 "k8s.io/api/core/v1"
)

// PersistentVolume represents a collection of PersistentVolumes available on a cluster.
type PersistentVolume struct {
	pvs map[string]*v1.PersistentVolume
}

// NewPersistentVolume returns a new PersistentVolume.
func NewPersistentVolume(pvs map[string]*v1.PersistentVolume) *PersistentVolume {
	return &PersistentVolume{pvs}
}

// ListPersistentVolumes returns all available PersistentVolumes on the cluster.
func (p *PersistentVolume) ListPersistentVolumes() map[string]*v1.PersistentVolume {
	return p.pvs
}
