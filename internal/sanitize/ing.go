package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	netv1b1 "k8s.io/api/networking/v1beta1"
)

type (
	// Ingress tracks Ingress sanitization.
	Ingress struct {
		*issues.Collector
		IngressLister
	}

	// IngLister list ingresses.
	IngLister interface {
		ListIngresses() map[string]*netv1b1.Ingress
	}

	// IngressLister list available Ingresss on a cluster.
	IngressLister interface {
		IngLister
	}
)

// NewIngress returns a new sanitizer.
func NewIngress(co *issues.Collector, lister IngressLister) *Ingress {
	return &Ingress{
		Collector:     co,
		IngressLister: lister,
	}
}

// Sanitize cleanse the resource.
func (i *Ingress) Sanitize(ctx context.Context) error {
	for fqn, ing := range i.ListIngresses() {
		i.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		i.checkDeprecation(ctx, ing)

		if i.NoConcerns(fqn) && i.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			i.ClearOutcome(fqn)
		}
	}

	return nil
}

func (i *Ingress) checkDeprecation(ctx context.Context, ing *netv1b1.Ingress) {
	const current = "networking.k8s.io/v1"
	rev, err := resourceRev(internal.MustExtractFQN(ctx), "Ingress", ing.Annotations)
	if err != nil {
		if rev = revFromLink(ing.SelfLink); rev == "" {
			return
		}
	}
	if rev != current {
		i.AddCode(ctx, 403, "Ingress", rev, current)
	}
}
