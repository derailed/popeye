package cache

import (
	netv1 "k8s.io/api/networking/v1"
	"sync"

	"github.com/derailed/popeye/internal"
)

// IngressKey tracks Ingress resource references
const IngressKey = "ing"

// Ingress represents Ingress cache.
type Ingress struct {
	ings map[string]*netv1.Ingress
}

// NewIngress returns a new Ingress cache.
func NewIngress(ings map[string]*netv1.Ingress) *Ingress {
	return &Ingress{ings: ings}
}

// ListIngresses returns all available Ingresss on the cluster.
func (d *Ingress) ListIngresses() map[string]*netv1.Ingress {
	return d.ings
}

// IngressRefs computes all ingress external references.
func (d *Ingress) IngressRefs(refs *sync.Map) {
	for _, ing := range d.ings {
		for _, tls := range ing.Spec.TLS {
			d.trackReference(refs, ResFqn(SecretKey, FQN(ing.Namespace, tls.SecretName)))
		}
	}
}

func (d *Ingress) trackReference(refs *sync.Map, key string) {
	refs.Store(key, internal.AllKeys)
}
