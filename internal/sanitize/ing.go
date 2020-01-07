package sanitize

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	nv1beta1 "k8s.io/api/extensions/v1beta1"
)

type (
	// Ingress tracks Ingress sanitization.
	Ingress struct {
		*issues.Collector
		IngressLister
	}

	// IngLister list ingresses.
	IngLister interface {
		ListIngresses() map[string]*nv1beta1.Ingress
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

		if i.Config.ExcludeFQN(internal.MustExtractSection(ctx), fqn) {
			i.ClearOutcome(fqn)
		}
	}

	return nil
}

func (i *Ingress) checkDeprecation(ctx context.Context, ing *nv1beta1.Ingress) {
	const current = "networking.k8s.io/v1beta1"

	rev, err := resourceRev(internal.MustExtractFQN(ctx), ing.Annotations)
	if err != nil {
		rev = revFromLink(ing.SelfLink)
		if rev == "" {
			i.AddCode(ctx, 404, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		i.AddCode(ctx, 403, "Ingress", rev, current)
	}
}
