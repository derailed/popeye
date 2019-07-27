package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// ReplicaSet represents a ReplicaSet sanitizer.
type ReplicaSet struct {
	*issues.Collector
	*cache.ReplicaSet
	*cache.Pod
	*config.Config

	client *k8s.Client
}

// NewReplicaSet return a new ReplicaSet sanitizer.
func NewReplicaSet(c *Cache, codes *issues.Codes) Sanitizer {
	d := ReplicaSet{
		client:    c.client,
		Config:    c.config,
		Collector: issues.NewCollector(codes),
	}

	dps, err := c.replicasets()
	if err != nil {
		d.AddErr("replicasets", err)
	}
	d.ReplicaSet = dps

	pod, err := c.pods()
	if err != nil {
		d.AddErr("pods", err)
	}
	d.Pod = pod

	return &d
}

// Sanitize all available ReplicaSets.
func (d *ReplicaSet) Sanitize(ctx context.Context) error {
	return sanitize.NewReplicaSet(d.Collector, d).Sanitize(ctx)
}
