// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type direction string

const (
	dirIn    direction = "Ingress"
	dirOut   direction = "Egress"
	bothPols           = "all"
	noPols             = ""
	ingress            = "ingress"
	egress             = "egress"
)

// NetworkPolicy tracks NetworkPolicy linting.
type NetworkPolicy struct {
	*issues.Collector

	db      *db.DB
	ipCache map[string][]v1.PodIP
}

// NewNetworkPolicy returns a new instance.
func NewNetworkPolicy(co *issues.Collector, db *db.DB) *NetworkPolicy {
	return &NetworkPolicy{
		Collector: co,
		db:        db,
		ipCache:   make(map[string][]v1.PodIP),
	}
}

// Lint cleanse the resource.
func (s *NetworkPolicy) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[internal.NP])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		np := o.(*netv1.NetworkPolicy)
		fqn := client.FQN(np.Namespace, np.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, np))

		s.checkSelector(ctx, fqn, np.Spec.PodSelector)
		s.checkIngresses(ctx, fqn, np.Spec.Ingress)
		s.checkEgresses(ctx, fqn, np.Spec.Egress)
		s.checkRuleType(ctx, &np.Spec)
	}

	return nil
}

func (s *NetworkPolicy) checkRuleType(ctx context.Context, spec *netv1.NetworkPolicySpec) {
	if spec.PodSelector.Size() > 0 {
		return
	}

	switch {
	case isAllowAll(spec):
		s.AddCode(ctx, 1203, "Allow", bothPols)
	case isAllowAllIngress(spec):
		s.AddCode(ctx, 1203, "Allow all", ingress)
	case isAllowAllEgress(spec):
		s.AddCode(ctx, 1203, "Allow all", egress)
	case isDenyAll(spec):
		s.AddCode(ctx, 1203, "Deny", bothPols)
	case isDenyAllIngress(spec):
		s.AddCode(ctx, 1203, "Deny all", ingress)
	case isDenyAllEgress(spec):
		s.AddCode(ctx, 1203, "Deny all", egress)
	}
}

func isDefaultDenyAll(np *netv1.NetworkPolicy) bool {
	if len(np.Spec.Ingress) > 0 {
		return false
	}
	if len(np.Spec.Egress) > 0 {
		return false
	}
	if np.Spec.PodSelector.Size() > 0 {
		return false
	}

	return len(np.Spec.PolicyTypes) == 2
}

func (s *NetworkPolicy) checkSelector(ctx context.Context, fqn string, sel metav1.LabelSelector) {
	ns, _ := client.Namespaced(fqn)
	if sel.Size() > 0 {
		pp, err := s.db.FindPodsBySel(ns, &sel)
		if err != nil || len(pp) == 0 {
			s.AddCode(ctx, 1200, dumpSel(&sel))
			return
		}
	}
}

func (s *NetworkPolicy) checkIngresses(ctx context.Context, fqn string, rr []netv1.NetworkPolicyIngressRule) {
	for _, r := range rr {
		for _, from := range r.From {
			s.checkSelectors(ctx, fqn, from.NamespaceSelector, from.PodSelector, dirIn)
			s.checkIPBlocks(ctx, fqn, from.IPBlock, dirIn)
		}
	}
}

func (s *NetworkPolicy) checkEgresses(ctx context.Context, fqn string, rr []netv1.NetworkPolicyEgressRule) {
	for _, r := range rr {
		for _, to := range r.To {
			s.checkSelectors(ctx, fqn, to.NamespaceSelector, to.PodSelector, dirOut)
			s.checkIPBlocks(ctx, fqn, to.IPBlock, dirOut)
		}
	}
}

func (s *NetworkPolicy) checkSelectors(ctx context.Context, fqn string, nsSel, podSel *metav1.LabelSelector, d direction) {
	ns, _ := client.Namespaced(fqn)
	if nsSel != nil && nsSel.Size() > 0 {
		nss, err := s.db.FindNSBySel(nsSel)
		if err != nil {
			s.AddErr(ctx, fmt.Errorf("unable to locate namespace using selector: %s", dumpSel(nsSel)))
			return
		}
		s.checkNSSelector(ctx, nsSel, nss, d)
		s.checkPodSelector(ctx, nss, podSel, d)
		return
	}
	nss, err := s.db.FindNS(ns)
	if err != nil {
		s.AddErr(ctx, fmt.Errorf("unable to locate namespace: %q", ns))
		return
	}
	s.checkPodSelector(ctx, []*v1.Namespace{nss}, podSel, d)
}

func (s *NetworkPolicy) checkIPBlocks(ctx context.Context, fqn string, b *netv1.IPBlock, d direction) {
	if b == nil {
		return
	}
	ns, _ := client.Namespaced(fqn)
	_, ipnet, err := net.ParseCIDR(b.CIDR)
	if err != nil {
		s.AddErr(ctx, err)
	}
	if !s.matchPips(ns, ipnet) {
		s.AddCode(ctx, 1206, strings.ToLower(string(d)), b.CIDR)
	}
	for _, ex := range b.Except {
		_, ipnet, err := net.ParseCIDR(ex)
		if err != nil {
			s.AddErr(ctx, err)
			continue
		}
		if !s.matchPips(ns, ipnet) {
			s.AddCode(ctx, 1207, strings.ToLower(string(d)), ex)
		}
	}
}

func (s *NetworkPolicy) matchPips(ns string, ipnet *net.IPNet) bool {
	if ipnet == nil {
		return false
	}
	txn, it := s.db.MustITForNS(internal.Glossary[internal.PO], ns)
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		po := o.(*v1.Pod)
		for _, ip := range po.Status.PodIPs {
			if ipnet.Contains(net.ParseIP(ip.IP)) {
				return true
			}
		}
	}

	return false
}

func (s *NetworkPolicy) checkPodSelector(ctx context.Context, nss []*v1.Namespace, sel *metav1.LabelSelector, d direction) {
	if sel == nil || sel.Size() == 0 {
		return
	}

	var found bool
	nn := make([]string, 0, len(nss))
	for _, ns := range nss {
		pp, err := s.db.FindPodsBySel(ns.Name, sel)
		if err != nil {
			s.AddErr(ctx, fmt.Errorf("unable to locate pods by selector: %w", err))
			return
		}
		if len(pp) > 0 {
			found = true
		} else {
			nn = append(nn, ns.Name)
		}
	}
	if !found {
		if len(nn) > 0 {
			s.AddCode(ctx, 1208, strings.ToLower(string(d)), dumpSel(sel), strings.Join(nn, ","))
		} else {
			s.AddCode(ctx, 1202, strings.ToLower(string(d)), dumpSel(sel))
		}
	}
}

func (s *NetworkPolicy) checkNSSelector(ctx context.Context, sel *metav1.LabelSelector, nss []*v1.Namespace, d direction) bool {
	if len(nss) == 0 {
		s.AddCode(ctx, 1201, strings.ToLower(string(d)), dumpSel(sel))
		return false
	}

	return true
}

// Helpers...

func dumpLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	ll := make([]string, 0, len(labels))
	for k, v := range labels {
		ll = append(ll, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(ll, ",")
}

func dumpSel(sel *metav1.LabelSelector) string {
	if sel == nil {
		return "n/a"
	}

	var out string
	out = dumpLabels(sel.MatchLabels)

	ll := make([]string, 0, len(sel.MatchExpressions))
	for _, v := range sel.MatchExpressions {
		ll = append(ll, fmt.Sprintf("%s-%s-%s", v.Key, v.Operator, strings.Join(v.Values, ",")))
	}
	if out != "" && len(ll) > 0 {
		out += "|"
	}
	out += strings.Join(ll, ",")

	return out
}
