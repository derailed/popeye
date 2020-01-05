package cache

import (
	nv1beta1 "k8s.io/api/extensions/v1beta1"
)

// IngressKey tracks Ingress resource references
const IngressKey = "ing"

// Ingress represents Ingress cache.
type Ingress struct {
	ings map[string]*nv1beta1.Ingress
}

// NewIngress returns a new Ingress cache.
func NewIngress(ings map[string]*nv1beta1.Ingress) *Ingress {
	return &Ingress{ings: ings}
}

// ListIngresses returns all available Ingresss on the cluster.
func (d *Ingress) ListIngresses() map[string]*nv1beta1.Ingress {
	return d.ings
}

// IngressRefs computes all ingress external references.
func (d *Ingress) IngressRefs(refs ObjReferences) {
	for _, ing := range d.ings {
		for _, tls := range ing.Spec.TLS {
			d.trackReference(refs, ResFqn(SecretKey, FQN(ing.Namespace, tls.SecretName)))
		}
	}
}

func (d *Ingress) trackReference(refs ObjReferences, key string) {
	if set, ok := refs[key]; ok {
		set.Add(AllKeys)
	} else {
		refs[key] = StringSet{AllKeys: Blank}
	}
}
