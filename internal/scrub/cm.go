package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// ConfigMap represents a configMap scruber.
type ConfigMap struct {
	*issues.Collector
	*cache.Pod
	*cache.ConfigMap
}

// NewConfigMap return a new ConfigMap scruber.
func NewConfigMap(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	s := ConfigMap{Collector: issues.NewCollector(codes, c.config)}

	var err error
	s.ConfigMap, err = c.configmaps()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.Pod, err = c.pods()
	if err != nil {
		s.AddErr(ctx, err)
	}

	return &s
}

// Sanitize all available ConfigMaps.
func (c *ConfigMap) Sanitize(ctx context.Context) error {
	return sanitize.NewConfigMap(c.Collector, c).Sanitize(ctx)
}
