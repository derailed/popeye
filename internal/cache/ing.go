package cache

import (
	extv1beta1 "k8s.io/api/extensions/v1beta1"
)

// IngressKey tracks Ingress ressource references
const IngressKey = "ing"

// Ingress represents Ingress cache.
type Ingress struct {
	ings map[string]*extv1beta1.Ingress
}

// NewIngress returns a new Ingress cache.
func NewIngress(ings map[string]*extv1beta1.Ingress) *Ingress {
	return &Ingress{ings: ings}
}

// ListIngresses returns all available Ingresss on the cluster.
func (d *Ingress) ListIngresses() map[string]*extv1beta1.Ingress {
	return d.ings
}
