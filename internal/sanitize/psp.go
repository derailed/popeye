package sanitize

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	pv1beta1 "k8s.io/api/policy/v1beta1"
)

type (
	// PodSecurityPolicy tracks PodSecurityPolicy sanitization.
	PodSecurityPolicy struct {
		*issues.Collector
		PodSecurityPolicyLister
	}

	// PodSecurityPolicyLister list available PodSecurityPolicys on a cluster.
	PodSecurityPolicyLister interface {
		ListPodSecurityPolicies() map[string]*pv1beta1.PodSecurityPolicy
	}
)

// NewPodSecurityPolicy returns a new sanitizer.
func NewPodSecurityPolicy(co *issues.Collector, lister PodSecurityPolicyLister) *PodSecurityPolicy {
	return &PodSecurityPolicy{
		Collector:               co,
		PodSecurityPolicyLister: lister,
	}
}

// Sanitize cleanse the resource.
func (p *PodSecurityPolicy) Sanitize(ctx context.Context) error {
	for fqn, psp := range p.ListPodSecurityPolicies() {
		p.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		p.checkDeprecation(ctx, psp)

		if p.Config.ExcludeFQN(internal.MustExtractSection(ctx), fqn) {
			p.ClearOutcome(fqn)
		}
	}

	return nil
}

func (p *PodSecurityPolicy) checkDeprecation(ctx context.Context, psp *pv1beta1.PodSecurityPolicy) {
	const current = "policy/v1beta1"

	rev, err := resourceRev(internal.MustExtractFQN(ctx), psp.Annotations)
	if err != nil {
		rev = revFromLink(psp.SelfLink)
		if rev == "" {
			p.AddCode(ctx, 404, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		p.AddCode(ctx, 403, "PodSecurityPolicy", rev, current)
	}
}
