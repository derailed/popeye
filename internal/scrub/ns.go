package scrub

import (
	"context"
	"fmt"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// Namespace represents a Namespace sanitizer.
type Namespace struct {
	*issues.Collector
	*cache.Namespace
	*cache.Pod
}

// NewNamespace return a new Namespace sanitizer.
func NewNamespace(c *k8s.Client, cfg *config.Config) Sanitizer {
	n := Namespace{Collector: issues.NewCollector()}

	ss, err := dag.ListNamespaces(c, cfg)
	if err != nil {
		n.AddErr("namespaces", err)
	}
	n.Namespace = cache.NewNamespace(ss)

	pods, err := dag.ListPods(c, cfg)
	if err != nil {
		n.AddErr("pods", err)
	}
	n.Pod = cache.NewPod(pods)

	return &n
}

// ReferencedNamespaces fetch all namespaces referenced by pods.
func (n *Namespace) ReferencedNamespaces(res map[string]struct{}) {
	refs := cache.ObjReferences{}
	n.Pod.PodRefs(refs)
	if nss, ok := refs["ns"]; ok {
		fmt.Println("NSS", nss)
		for ns := range nss {
			res[ns] = struct{}{}
		}
	}
}

// Sanitize all available Namespaces.
func (n *Namespace) Sanitize(ctx context.Context) error {
	return sanitize.NewNamespace(n.Collector, n).Sanitize(ctx)
}
