package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// Namespace represents a Namespace sanitizer.
type Namespace struct {
	*issues.Collector
	*cache.Namespace
	*cache.Pod
}

// NewNamespace return a new Namespace sanitizer.
func NewNamespace(c *Cache, codes *issues.Codes) Sanitizer {
	n := Namespace{Collector: issues.NewCollector(codes)}

	ss, err := dag.ListNamespaces(c.client, c.config)
	if err != nil {
		n.AddErr("namespaces", err)
	}
	n.Namespace = cache.NewNamespace(ss)

	pod, err := c.pods()
	if err != nil {
		n.AddErr("pods", err)
	}
	n.Pod = pod

	return &n
}

// ReferencedNamespaces fetch all namespaces referenced by pods.
func (n *Namespace) ReferencedNamespaces(res map[string]struct{}) {
	refs := cache.ObjReferences{}
	n.Pod.PodRefs(refs)
	if nss, ok := refs["ns"]; ok {
		for ns := range nss {
			res[ns] = struct{}{}
		}
	}
}

// Sanitize all available Namespaces.
func (n *Namespace) Sanitize(ctx context.Context) error {
	return sanitize.NewNamespace(n.Collector, n).Sanitize(ctx)
}
