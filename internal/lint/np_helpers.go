// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import netv1 "k8s.io/api/networking/v1"

func noPodSel(spec *netv1.NetworkPolicySpec) bool {
	return spec.PodSelector.Size() == 0
}

func isAllowAll(spec *netv1.NetworkPolicySpec) bool {
	return noPodSel(spec) &&
		blankIngress(spec.Ingress) && blankEgress(spec.Egress) &&
		len(spec.PolicyTypes) == 2
}

func isAllowAllIngress(spec *netv1.NetworkPolicySpec) bool {
	return noPodSel(spec) &&
		blankIngress(spec.Ingress) && polInclude(spec.PolicyTypes, dirIn)
}

func isAllowAllEgress(spec *netv1.NetworkPolicySpec) bool {
	return noPodSel(spec) &&
		blankEgress(spec.Egress) && polInclude(spec.PolicyTypes, dirOut)
}

func isDeny(spec *netv1.NetworkPolicySpec) bool {
	return noPodSel(spec) && spec.Egress == nil && spec.Ingress == nil
}

func isDenyAll(spec *netv1.NetworkPolicySpec) bool {
	return isDeny(spec) && len(spec.PolicyTypes) == 2
}

func isDenyAllIngress(spec *netv1.NetworkPolicySpec) bool {
	return noPodSel(spec) && spec.Ingress == nil && polInclude(spec.PolicyTypes, dirIn)
}

func isDenyAllEgress(spec *netv1.NetworkPolicySpec) bool {
	return noPodSel(spec) && spec.Egress == nil && polInclude(spec.PolicyTypes, dirOut)
}

func blankEgress(rr []netv1.NetworkPolicyEgressRule) bool {
	return len(rr) == 1 && len(rr[0].Ports) == 0 && len(rr[0].To) == 0
}

func blankIngress(rr []netv1.NetworkPolicyIngressRule) bool {
	return len(rr) == 1 && len(rr[0].Ports) == 0 && len(rr[0].From) == 0
}

func polInclude(pp []netv1.PolicyType, d direction) bool {
	for _, p := range pp {
		if p == netv1.PolicyType(d) {
			return true
		}
	}

	return false
}
