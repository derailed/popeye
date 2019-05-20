package cache

import (
	v1 "k8s.io/api/core/v1"
)

// Endpoints represents Endpoints cache.
type Endpoints struct {
	eps map[string]*v1.Endpoints
}

// NewEndpoints returns a new Endpoints cache.
func NewEndpoints(eps map[string]*v1.Endpoints) *Endpoints {
	return &Endpoints{eps: eps}
}

// GetEndpoints returns all available Endpoints on the cluster.
func (e *Endpoints) GetEndpoints(fqn string) *v1.Endpoints {
	return e.eps[fqn]
}
