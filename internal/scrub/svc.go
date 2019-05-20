package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// Service represents a Service sanitizer.
type Service struct {
	*issues.Collector
	*cache.Service
	*cache.Pod
	*cache.Endpoints
}

// NewService return a new Service sanitizer.
func NewService(c *k8s.Client, cfg *config.Config) Sanitizer {
	p := Service{Collector: issues.NewCollector()}

	ss, err := dag.ListServices(c, cfg)
	if err != nil {
		p.AddErr("services", err)
	}
	p.Service = cache.NewService(ss)

	pp, err := dag.ListPods(c, cfg)
	if err != nil {
		p.AddErr("pod", err)
	}
	p.Pod = cache.NewPod(pp)

	ee, err := dag.ListEndpoints(c, cfg)
	if err != nil {
		p.AddErr("endpoints", err)
	}
	p.Endpoints = cache.NewEndpoints(ee)

	return &p
}

// Sanitize all available Services.
func (s *Service) Sanitize(ctx context.Context) error {
	return sanitize.NewService(s.Collector, s).Sanitize(ctx)
}
