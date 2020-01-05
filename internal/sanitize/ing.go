package sanitize

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/issues"
	nv1beta1 "k8s.io/api/extensions/v1beta1"
)

type (
	// Ingress tracks Ingress sanitization.
	Ingress struct {
		*issues.Collector
		IngressLister
	}

	// IngLister list deployments.
	IngLister interface {
		ListIngresses() map[string]*nv1beta1.Ingress
	}

	// IngressLister list available Ingresss on a cluster.
	IngressLister interface {
		IngLister
	}
)

// NewIngress returns a new Ingress sanitizer.
func NewIngress(co *issues.Collector, lister IngressLister) *Ingress {
	return &Ingress{
		Collector:     co,
		IngressLister: lister,
	}
}

// Sanitize configmaps.
func (d *Ingress) Sanitize(ctx context.Context) error {
	for fqn, ing := range d.ListIngresses() {
		d.InitOutcome(fqn)
		d.checkDeprecation(fqn, ing)
	}

	return nil
}

func (d *Ingress) checkDeprecation(fqn string, ing *nv1beta1.Ingress) {
	const current = "networking.k8s.io/v1beta1"

	rev, err := resourceRev(fqn, ing.Annotations)
	if err != nil {
		rev = revFromLink(ing.SelfLink)
		if rev == "" {
			d.AddCode(404, fqn, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		d.AddCode(403, fqn, "Ingress", rev, current)
	}
}
