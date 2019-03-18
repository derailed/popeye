package linter

import (
	v1 "k8s.io/api/core/v1"
)

// Check port mappings
// Check endpoints
// Check LoadBalancer type

// Service represents a service linter.
type Service struct {
	*Linter
}

// NewService returns a new service linter.
func NewService() *Service {
	return &Service{new(Linter)}
}

// Lint a service.
func (*Service) Lint(s v1.Service) {
}
