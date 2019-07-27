package cache

import (
	pv1beta1 "k8s.io/api/extensions/v1beta1"
)

// PodSecurityPolicyKey tracks PodSecurityPolicy ressource references
const PodSecurityPolicyKey = "psp"

// PodSecurityPolicy represents PodSecurityPolicy cache.
type PodSecurityPolicy struct {
	psps map[string]*pv1beta1.PodSecurityPolicy
}

// NewPodSecurityPolicy returns a new PodSecurityPolicy cache.
func NewPodSecurityPolicy(psps map[string]*pv1beta1.PodSecurityPolicy) *PodSecurityPolicy {
	return &PodSecurityPolicy{psps: psps}
}

// ListPodSecurityPolicies returns all available PodSecurityPolicies on the cluster.
func (p *PodSecurityPolicy) ListPodSecurityPolicies() map[string]*pv1beta1.PodSecurityPolicy {
	return p.psps
}
