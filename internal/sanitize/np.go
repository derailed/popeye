package sanitize

import (
	"context"
	"errors"

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

		if n.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			n.ClearOutcome(fqn)
		}
	}

	return nil
}

func (n *NetworkPolicy) checkPodSelector(ctx context.Context, ns string, sel *metav1.LabelSelector, kind string) {
	if sel == nil {
		return
	}

	if pods := n.ListPodsBySelector(ns, sel); len(pods) == 0 {
		n.AddCode(ctx, 1200, kind)
	}
}

func (n *NetworkPolicy) checkNSSelector(ctx context.Context, sel *metav1.LabelSelector, kind string) {
	if sel == nil {
		return
	}

	if nss := n.ListNamespacesBySelector(sel); len(nss) == 0 {
		n.AddCode(ctx, 1201, kind)
	}
}

func (n *NetworkPolicy) checkRefs(ctx context.Context, np *nv1.NetworkPolicy) {
	for _, ing := range np.Spec.Ingress {
		for _, f := range ing.From {
			n.checkPodSelector(ctx, np.Namespace, f.PodSelector, "Ingress")
			n.checkNSSelector(ctx, f.NamespaceSelector, "Ingress")
		}
	}

	for _, eg := range np.Spec.Egress {
		for _, f := range eg.To {
			n.checkPodSelector(ctx, np.Namespace, f.PodSelector, "Egress")
			n.checkNSSelector(ctx, f.NamespaceSelector, "Egress")
		}
	}
}

func (n *NetworkPolicy) checkDeprecation(ctx context.Context, np *nv1.NetworkPolicy) {
	const current = "networking.k8s.io/v1"

	rev, err := resourceRev(internal.MustExtractFQN(ctx), "NetworkPolicy", np.Annotations)
	if err != nil {
		rev = revFromLink(np.SelfLink)
		if rev == "" {
			n.AddCode(ctx, 404, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		n.AddCode(ctx, 403, "NetworkPolicy", rev, current)
	}
}
