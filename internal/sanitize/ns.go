package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

type (
	// NamespaceLister lists all namespaces.
	NamespaceLister interface {
		NamespaceRefs
		ListNamespaces() map[string]*v1.Namespace
	}

	// NamespaceRefs tracks namespace references in the cluster.
	NamespaceRefs interface {
		ReferencedNamespaces(map[string]struct{})
	}

	// Namespace represents a Namespace sanitizer.
	Namespace struct {
		*issues.Collector
		NamespaceLister
	}
)

// NewNamespace returns a new sanitizer.
func NewNamespace(co *issues.Collector, lister NamespaceLister) *Namespace {
	return &Namespace{
		Collector:       co,
		NamespaceLister: lister,
	}
}

// Sanitize cleanse the resource.
func (n *Namespace) Sanitize(ctx context.Context) error {
	available := n.ListNamespaces()
	used := make(map[string]struct{}, len(available))
	n.ReferencedNamespaces(used)
	for fqn, ns := range available {
		n.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)
		if n.checkActive(ctx, ns.Status.Phase) {
			if _, ok := used[fqn]; !ok {
				n.AddCode(ctx, 400)
			}
		}
		if n.NoConcerns(fqn) && n.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			n.ClearOutcome(fqn)
		}
	}

	return nil
}

func (n *Namespace) checkActive(ctx context.Context, p v1.NamespacePhase) bool {
	if !isNSActive(p) {
		n.AddCode(ctx, 800)
		return false
	}

	return true
}

// ----------------------------------------------------------------------------
// Helpers...

func isNSActive(phase v1.NamespacePhase) bool {
	return phase == v1.NamespaceActive
}
