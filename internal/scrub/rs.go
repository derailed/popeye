package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
)

// ReplicaSet represents a ReplicaSet scruber.
type ReplicaSet struct {
	*issues.Collector
	*cache.ReplicaSet
	*cache.Pod
	*config.Config

	client types.Connection
}

// NewReplicaSet return a new ReplicaSet scruber.
func NewReplicaSet(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	d := ReplicaSet{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	d.ReplicaSet, err = c.replicasets()
	if err != nil {
		d.AddErr(ctx, err)
	}

	d.Pod, err = c.pods()
	if err != nil {
		d.AddErr(ctx, err)
	}

	return &d
}

// Sanitize all available ReplicaSets.
func (d *ReplicaSet) Sanitize(ctx context.Context) error {
	return sanitize.NewReplicaSet(d.Collector, d).Sanitize(ctx)
}
