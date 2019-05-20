package cache

import (
	v1 "k8s.io/api/core/v1"
)

// Service represents a collection of Services available on a cluster.
type Service struct {
	svcs map[string]*v1.Service
}

// NewService returns a new Service.
func NewService(svcs map[string]*v1.Service) *Service {
	return &Service{svcs}
}

// ListServices returns all available Services on the cluster.
func (s *Service) ListServices() map[string]*v1.Service {
	return s.svcs
}
