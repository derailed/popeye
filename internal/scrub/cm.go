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
func NewConfigMap(c *k8s.Client, cfg *config.Config) Sanitizer {
	s := ConfigMap{Collector: issues.NewCollector()}

	cms, err := dag.ListConfigMaps(c, cfg)
	if err != nil {
		s.AddErr("configmaps", err)
	}
	pods, err := dag.ListPods(c, cfg)
	if err != nil {
		s.AddErr("pods", err)
	}
	s.ConfigMap, s.Pod = cache.NewConfigMap(cms), cache.NewPod(pods)

	return &s
}

// Sanitize all available ConfigMaps.
func (c *ConfigMap) Sanitize(ctx context.Context) error {
	return sanitize.NewConfigMap(c.Collector, c).Sanitize(ctx)
}
