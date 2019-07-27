package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

// SkipNamespaces excludes system namespaces with no pods from being included in scan.
// BOZO!! spinachyaml default??
var skipNamespaces = []string{"default", "kube-public", "kube-node-lease"}

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

// NewNamespace returns a new namespace linter.
func NewNamespace(co *issues.Collector, lister NamespaceLister) *Namespace {
	return &Namespace{
		Collector:       co,
		NamespaceLister: lister,
	}
}

// Sanitize a namespace
func (n *Namespace) Sanitize(ctx context.Context) error {
	available := n.ListNamespaces()
	used := make(map[string]struct{}, len(available))
	n.ReferencedNamespaces(used)
	for fqn, ns := range available {
		n.InitOutcome(fqn)
		if n.checkActive(fqn, ns.Status.Phase) {
			if _, ok := used[fqn]; !ok {
				n.AddCode(400, fqn)
			}
		}
	}

	return nil
}

func (n *Namespace) checkActive(fqn string, p v1.NamespacePhase) bool {
	if !isNSActive(p) {
		n.AddCode(800, fqn)
		return false
	}

	return true
}

// ----------------------------------------------------------------------------
// Helpers...

func isNSActive(phase v1.NamespacePhase) bool {
	return phase == v1.NamespaceActive
}
