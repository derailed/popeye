package sanitize

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
	nv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	// NetworkPolicy tracks NetworkPolicy sanitization.
	NetworkPolicy struct {
		*issues.Collector
		NetworkPolicyLister
	}

	// NamespaceSelectorLister list a collection of namespaces matching a selector.
	NamespaceSelectorLister interface {
		ListNamespacesBySelector(sel *metav1.LabelSelector) map[string]*v1.Namespace
	}

	// NetworkPolicyLister list available NetworkPolicys on a cluster.
	NetworkPolicyLister interface {
		PodSelectorLister
		NamespaceSelectorLister
		ConfigLister
		ListNetworkPolicies() map[string]*nv1.NetworkPolicy
	}
)

// NewNetworkPolicy returns a new NetworkPolicy sanitizer.
func NewNetworkPolicy(co *issues.Collector, lister NetworkPolicyLister) *NetworkPolicy {
	return &NetworkPolicy{
		Collector:           co,
		NetworkPolicyLister: lister,
	}
}

// Sanitize configmaps.
func (n *NetworkPolicy) Sanitize(ctx context.Context) error {
	for fqn, np := range n.ListNetworkPolicies() {
		n.InitOutcome(fqn)
		n.checkDeprecation(fqn, np)
		n.checkRefs(fqn, np)
	}

	return nil
}

func (n *NetworkPolicy) checkRefs(fqn string, np *nv1.NetworkPolicy) {
	for _, ing := range np.Spec.Ingress {
		for _, f := range ing.From {
			if f.PodSelector != nil {
				pods := n.ListPodsBySelector(f.PodSelector)
				if len(pods) == 0 {
					n.AddCode(1200, fqn, "Ingress", f.PodSelector.String())
				}
			}
			if f.NamespaceSelector != nil {
				nss := n.ListNamespacesBySelector(f.NamespaceSelector)
				if len(nss) == 0 {
					n.AddCode(1201, fqn, "Ingress", f.NamespaceSelector.String())
				}
			}
		}
	}

	for _, eg := range np.Spec.Egress {
		for _, f := range eg.To {
			if f.PodSelector != nil {
				pods := n.ListPodsBySelector(f.PodSelector)
				if len(pods) == 0 {
					n.AddCode(1200, fqn, "Egress", f.PodSelector.String())
				}
			}
			if f.NamespaceSelector != nil {
				nss := n.ListNamespacesBySelector(f.NamespaceSelector)
				if len(nss) == 0 {
					n.AddCode(1201, fqn, "Egress", f.NamespaceSelector.String())
				}
			}
		}
	}
}

func (n *NetworkPolicy) checkDeprecation(fqn string, np *nv1.NetworkPolicy) {
	const current = "networking.k8s.io/v1"

	rev, err := resourceRev(fqn, np.Annotations)
	if err != nil {
		rev = revFromLink(np.SelfLink)
		if rev == "" {
			n.AddCode(404, fqn, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		n.AddCode(403, fqn, "NetworkPolicy", rev, current)
	}
}

// Helpers...
