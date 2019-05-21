package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// ConfigMap represents a configMap sanitizer.
type ConfigMap struct {
	*issues.Collector
	*cache.Pod
	*cache.ConfigMap
}

// Sanitizer represents a resource sanitizer.
type Sanitizer interface {
	Collector
	Sanitize(context.Context) error
}

// Collector collects sanitization issues.
type Collector interface {
	MaxSeverity(res string) issues.Level
	Outcome() issues.Outcome
}

// NewConfigMap return a new ConfigMap sanitizer.
func NewConfigMap(c *Cache) Sanitizer {
	s := ConfigMap{Collector: issues.NewCollector()}

	cms, err := dag.ListConfigMaps(c.client, c.config)
	if err != nil {
		s.AddErr("configmaps", err)
	}
	s.ConfigMap = cache.NewConfigMap(cms)

	pod, err := c.pods()
	if err != nil {
		s.AddErr("pods", err)
	}
	s.Pod = pod

	return &s
}

// Sanitize all available ConfigMaps.
func (c *ConfigMap) Sanitize(ctx context.Context) error {
	return sanitize.NewConfigMap(c.Collector, c).Sanitize(ctx)
}
