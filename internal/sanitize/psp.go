package sanitize

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/issues"
	pv1beta1 "k8s.io/api/extensions/v1beta1"
)

type (
	// PodSecurityPolicy tracks PodSecurityPolicy sanitization.
	PodSecurityPolicy struct {
		*issues.Collector
		PodSecurityPolicyLister
	}

	// PodSecurityPolicyLister list available PodSecurityPolicys on a cluster.
	PodSecurityPolicyLister interface {
		ConfigLister
		ListPodSecurityPolicies() map[string]*pv1beta1.PodSecurityPolicy
	}
)

// NewPodSecurityPolicy returns a new PodSecurityPolicy sanitizer.
func NewPodSecurityPolicy(co *issues.Collector, lister PodSecurityPolicyLister) *PodSecurityPolicy {
	return &PodSecurityPolicy{
		Collector:               co,
		PodSecurityPolicyLister: lister,
	}
}

// Sanitize configmaps.
func (p *PodSecurityPolicy) Sanitize(ctx context.Context) error {
	for fqn, psp := range p.ListPodSecurityPolicies() {
		p.InitOutcome(fqn)
		p.checkDeprecation(fqn, psp)
	}

	return nil
}

func (p *PodSecurityPolicy) checkDeprecation(fqn string, psp *pv1beta1.PodSecurityPolicy) {
	const current = "policy/v1beta1"

	rev, err := resourceRev(fqn, psp.Annotations)
	if err != nil {
		rev = revFromLink(psp.SelfLink)
		if rev == "" {
			p.AddCode(404, fqn, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		p.AddCode(403, fqn, "PodSecurityPolicy", rev, current)
	}
}

// Helpers...
