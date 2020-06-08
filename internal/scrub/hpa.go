package scrub

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// HorizontalPodAutoscaler represents a HorizontalPodAutoscaler scruber.
type HorizontalPodAutoscaler struct {
	*issues.Collector
	*cache.HorizontalPodAutoscaler
	*cache.Pod
	*cache.Node
	*cache.PodsMetrics
	*cache.NodesMetrics
	*cache.Deployment
	*cache.StatefulSet
	*cache.ServiceAccount
	*config.Config
}

// NewHorizontalPodAutoscaler return a new HorizontalPodAutoscaler scruber.
func NewHorizontalPodAutoscaler(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	h := HorizontalPodAutoscaler{
		Collector: issues.NewCollector(codes, c.config),
		Config:    c.config,
	}

	ctx = context.WithValue(ctx, internal.KeyFactory, c.factory)
	ctx = context.WithValue(ctx, internal.KeyConfig, c.config)
	ctx = context.WithValue(ctx, internal.KeyConfig, c.config)
	if c.config.Flags.ActiveNamespace != nil {
		ctx = context.WithValue(ctx, internal.KeyNamespace, *c.config.Flags.ActiveNamespace)
	} else {
		ns, err := c.factory.Client().Config().CurrentNamespaceName()
		if err != nil {
			ns = client.AllNamespaces
		}
		ctx = context.WithValue(ctx, internal.KeyNamespace, ns)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var err error
	ss, err := dag.ListHorizontalPodAutoscalers(ctx)
	if err != nil {
		h.AddErr(ctx, err)
	}
	h.HorizontalPodAutoscaler = cache.NewHorizontalPodAutoscaler(ss)

	h.Deployment, err = c.deployments()
	if err != nil {
		h.AddErr(ctx, err)
	}

	h.StatefulSet, err = c.statefulsets()
	if err != nil {
		h.AddErr(ctx, err)
	}

	h.Node, err = c.nodes()
	if err != nil {
		h.AddCode(ctx, 402, err)
	}

	h.NodesMetrics, _ = c.nodesMx()

	h.Pod, err = c.pods()
	if err != nil {
		h.AddErr(ctx, err)
	}
	h.PodsMetrics, _ = c.podsMx()

	h.ServiceAccount, err = c.serviceaccounts()
	if err != nil {
		h.AddErr(ctx, err)
	}

	return &h
}

// Sanitize all available HorizontalPodAutoscalers.
func (h *HorizontalPodAutoscaler) Sanitize(ctx context.Context) error {
	return sanitize.NewHorizontalPodAutoscaler(h.Collector, h).Sanitize(ctx)
}
