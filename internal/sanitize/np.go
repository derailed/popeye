package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
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
		ListNetworkPolicies() map[string]*nv1.NetworkPolicy
	}
)

// NewNetworkPolicy returns a new sanitizer.
func NewNetworkPolicy(co *issues.Collector, lister NetworkPolicyLister) *NetworkPolicy {
	return &NetworkPolicy{
		Collector:           co,
		NetworkPolicyLister: lister,
	}
}

// Sanitize cleanse the resource.
func (n *NetworkPolicy) Sanitize(ctx context.Context) error {
	for fqn, np := range n.ListNetworkPolicies() {
		n.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		n.checkDeprecation(ctx, np)
		n.checkRefs(ctx, np)

		if n.NoConcerns(fqn) && n.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			n.ClearOutcome(fqn)
		}
	}

	return nil
}

func (n *NetworkPolicy) checkPodSelector(ctx context.Context, nss map[string]*v1.Namespace, sel *metav1.LabelSelector, kind string) {
	if sel == nil {
		return
	}

	var found bool
	for ns := range nss {
		if pods := n.ListPodsBySelector(ns, sel); len(pods) > 0 {
			found = true
		}
	}
	if !found {
		n.AddCode(ctx, 1200, kind)
	}
}

func (n *NetworkPolicy) checkNSSelector(ctx context.Context, sel *metav1.LabelSelector, kind string) map[string]*v1.Namespace {
	if sel == nil {
		return nil
	}

	nss := n.ListNamespacesBySelector(sel)
	if len(nss) == 0 {
		n.AddCode(ctx, 1201, kind)
	}

	return nss
}

func (n *NetworkPolicy) checkRefs(ctx context.Context, np *nv1.NetworkPolicy) {
	const (
		ingress = "Ingress"
		egress  = "Egress"
	)

	for _, ing := range np.Spec.Ingress {
		for _, from := range ing.From {
			if from.NamespaceSelector != nil {
				if nss := n.checkNSSelector(ctx, from.NamespaceSelector, ingress); len(nss) > 0 {
					n.checkPodSelector(ctx, nss, from.PodSelector, ingress)
				}
			} else {
				n.checkPodSelector(ctx, map[string]*v1.Namespace{np.Namespace: nil}, from.PodSelector, ingress)
			}
		}
	}

	for _, eg := range np.Spec.Egress {
		for _, to := range eg.To {
			if to.NamespaceSelector != nil {
				if nss := n.checkNSSelector(ctx, to.NamespaceSelector, egress); len(nss) > 0 {
					n.checkPodSelector(ctx, nss, to.PodSelector, egress)
				}
			} else {
				n.checkPodSelector(ctx, map[string]*v1.Namespace{np.Namespace: nil}, to.PodSelector, egress)
			}
		}
	}
}

func (n *NetworkPolicy) checkDeprecation(ctx context.Context, np *nv1.NetworkPolicy) {
	const current = "networking.k8s.io/v1"

	rev, err := resourceRev(internal.MustExtractFQN(ctx), "NetworkPolicy", np.Annotations)
	if err != nil {
		if rev = revFromLink(np.SelfLink); rev == "" {
			return
		}
	}
	if rev != current {
		n.AddCode(ctx, 403, "NetworkPolicy", rev, current)
	}
}
