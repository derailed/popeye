package cache

import (
	v1 "k8s.io/api/core/v1"
)

// PersistentVolumeClaim represents a collection of PersistentVolumeClaims available on a cluster.
type PersistentVolumeClaim struct {
	pvcs map[string]*v1.PersistentVolumeClaim
}

// NewPersistentVolumeClaim returns a new PersistentVolumeClaim.
func NewPersistentVolumeClaim(pvcs map[string]*v1.PersistentVolumeClaim) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{pvcs}
}

// ListPersistentVolumeClaims returns all available PersistentVolumeClaims on the cluster.
func (p *PersistentVolumeClaim) ListPersistentVolumeClaims() map[string]*v1.PersistentVolumeClaim {
	return p.pvcs
}
