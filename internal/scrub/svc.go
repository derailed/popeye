package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// Service represents a Service sanitizer.
type Service struct {
	*issues.Collector
	*cache.Service
	*cache.Pod
	*cache.Endpoints
}

// NewService return a new Service sanitizer.
func NewService(c *Cache, codes *issues.Codes) Sanitizer {
	s := Service{Collector: issues.NewCollector(codes)}

	ss, err := dag.ListServices(c.client, c.config)
	if err != nil {
		s.AddErr("services", err)
	}
	s.Service = cache.NewService(ss)

	pod, err := c.pods()
	if err != nil {
		s.AddErr("pods", err)
	}
	s.Pod = pod

	ee, err := dag.ListEndpoints(c.client, c.config)
	if err != nil {
		s.AddErr("endpoints", err)
	}
	s.Endpoints = cache.NewEndpoints(ee)

	return &s
}

// Sanitize all available Services.
func (s *Service) Sanitize(ctx context.Context) error {
	return sanitize.NewService(s.Collector, s).Sanitize(ctx)
}
