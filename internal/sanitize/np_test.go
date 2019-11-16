package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	nv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNPSanitize(t *testing.T) {
	uu := map[string]struct {
		lister NetworkPolicyLister
		issues issues.Issues
	}{
		"good": {
			lister: makeNPLister(npOpts{
				rev: "networking.k8s.io/v1",
			}),
			issues: issues.Issues{},
		},
		"deprecated": {
			lister: makeNPLister(npOpts{
				rev: "policy/v1beta1",
			}),
			issues: issues.Issues{
				issues.Issue{
					Group:   "__root__",
					Level:   2,
					Message: `[POP-403] Deprecated NetworkPolicy API group "policy/v1beta1". Use "networking.k8s.io/v1" instead`},
			},
		},
		"noPodRef": {
			lister: makeNPLister(npOpts{
				rev: "networking.k8s.io/v1",
				pod: true,
			}),
			issues: issues.Issues{
				issues.Issue{
					Group:   "__root__",
					Level:   2,
					Message: "[POP-1200] No pods match Ingress pod selector",
				},
				issues.Issue{
					Group:   "__root__",
					Level:   2,
					Message: "[POP-1200] No pods match Egress pod selector",
				},
			},
		},
		"noNSRef": {
			lister: makeNPLister(npOpts{
				rev: "networking.k8s.io/v1",
				ns:  true,
			}),
			issues: issues.Issues{
				issues.Issue{
					Group:   "__root__",
					Level:   2,
					Message: "[POP-1201] No namespaces match Ingress namespace selector",
				},
				issues.Issue{
					Group:   "__root__",
					Level:   2,
					Message: "[POP-1201] No namespaces match Egress namespace selector",
				},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			np := NewNetworkPolicy(issues.NewCollector(loadCodes(t)), u.lister)

			assert.Nil(t, np.Sanitize(context.Background()))
			assert.Equal(t, u.issues, np.Outcome()["default/np"])
		})
	}
}

type (
	npOpts struct {
		rev     string
		pod, ns bool
	}

	np struct {
		name string
		opts npOpts
	}
)

func makeNPLister(opts npOpts) *np {
	return &np{
		name: "np",
		opts: opts,
	}
}

func (n *np) ListNetworkPolicies() map[string]*nv1.NetworkPolicy {
	return map[string]*nv1.NetworkPolicy{
		cache.FQN("default", n.name): makeNP(n.name, n.opts),
	}
}

func (n *np) ListNamespacesBySelector(sel *metav1.LabelSelector) map[string]*v1.Namespace {
	if n.opts.ns {
		return map[string]*v1.Namespace{}
	}

	return map[string]*v1.Namespace{
		"ns1": makeNS("ns1", true),
	}
}

func (n *np) ListPodsBySelector(sel *metav1.LabelSelector) map[string]*v1.Pod {
	if n.opts.pod {
		return map[string]*v1.Pod{}
	}

	return map[string]*v1.Pod{
		"default/p1": makePod("p1"),
	}
}

func (n *np) ListPods() map[string]*v1.Pod {
	return map[string]*v1.Pod{
		"default/p1": makePodSa("p1", "fred"),
	}
}

func (n *np) GetPod(map[string]string) *v1.Pod {
	return nil
}

func makeNP(n string, o npOpts) *nv1.NetworkPolicy {
	return &nv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
			SelfLink:  "/api/" + o.rev,
		},
		Spec: nv1.NetworkPolicySpec{
			Ingress: []nv1.NetworkPolicyIngressRule{
				{
					From: []nv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"po": "po1",
								},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"ns": "ns1",
								},
							},
						},
					},
				},
			},
			Egress: []nv1.NetworkPolicyEgressRule{
				{
					To: []nv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"po": "po1",
								},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"ns": "ns1",
								},
							},
						},
					},
				},
			},
		},
	}
}
