package cache

import (
	nv1 "k8s.io/api/networking/v1"
)

// NetworkPolicyKey tracks NetworkPolicy resource references
const NetworkPolicyKey = "np"

// NetworkPolicy represents NetworkPolicy cache.
type NetworkPolicy struct {
	nps map[string]*nv1.NetworkPolicy
}

// NewNetworkPolicy returns a new NetworkPolicy cache.
func NewNetworkPolicy(nps map[string]*nv1.NetworkPolicy) *NetworkPolicy {
	return &NetworkPolicy{nps: nps}
}

// ListNetworkPolicies returns all available NetworkPolicys on the cluster.
func (d *NetworkPolicy) ListNetworkPolicies() map[string]*nv1.NetworkPolicy {
	return d.nps
}
