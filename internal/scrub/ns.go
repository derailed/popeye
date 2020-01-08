package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// Namespace represents a Namespace scruber.
type Namespace struct {
	*issues.Collector
	*cache.Namespace
	*cache.Pod
}

// NewNamespace return a new Namespace scruber.
func NewNamespace(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	n := Namespace{Collector: issues.NewCollector(codes, c.config)}

	var err error
	n.Namespace, err = c.namespaces()
	if err != nil {
		n.AddErr(ctx, err)
	}

	n.Pod, err = c.pods()
	if err != nil {
		n.AddErr(ctx, err)
	}

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
