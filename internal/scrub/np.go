package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// NetworkPolicy represents a NetworkPolicy sanitizer.
type NetworkPolicy struct {
	*issues.Collector
	*cache.NetworkPolicy
	*cache.Namespace
	*cache.Pod
	*config.Config

	client *k8s.Client
}

// NewNetworkPolicy return a new NetworkPolicy sanitizer.
func NewNetworkPolicy(c *Cache, codes *issues.Codes) Sanitizer {
	n := NetworkPolicy{
		client:    c.client,
		Config:    c.config,
		Collector: issues.NewCollector(codes),
	}

	nps, err := c.networkpolicies()
	if err != nil {
		n.AddErr("networkpolicy", err)
	}
	n.NetworkPolicy = nps

	nss, err := c.namespaces()
	if err != nil {
		n.AddCode(402, "namespaces", err)
	}
	n.Namespace = nss

	pod, err := c.pods()
	if err != nil {
		n.AddErr("pods", err)
	}
	n.Pod = pod

	return &n
}

// Sanitize all available NetworkPolicys.
func (n *NetworkPolicy) Sanitize(ctx context.Context) error {
	return sanitize.NewNetworkPolicy(n.Collector, n).Sanitize(ctx)
}
