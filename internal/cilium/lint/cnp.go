// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"fmt"
	"strings"

	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	slimv1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/meta/v1"
	"github.com/cilium/cilium/pkg/policy/api"
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cilium"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	ilint "github.com/derailed/popeye/internal/lint"
)

// CiliumNetworkPolicy tracks CiliumNetworkPolicy sanitization.
type CiliumNetworkPolicy struct {
	*issues.Collector
	db *db.DB
}

// NewCiliumNetworkPolicy returns a new instance.
func NewCiliumNetworkPolicy(c *issues.Collector, db *db.DB) *CiliumNetworkPolicy {
	return &CiliumNetworkPolicy{
		Collector: c,
		db:        db,
	}
}

// Lint lints the resource.
func (s *CiliumNetworkPolicy) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[cilium.CNP])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		cnp := o.(*v2.CiliumNetworkPolicy)
		fqn := client.FQN(cnp.Namespace, cnp.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, ilint.SpecFor(fqn, cnp))

		rules := cnp.Specs
		if cnp.Spec != nil {
			rules = append(rules, cnp.Spec)
		}
		for _, r := range rules {
			if err := s.checkRule(ctx, cnp.Namespace, r); err != nil {
				s.AddErr(ctx, err)
			}
		}
	}

	return nil
}

func (s *CiliumNetworkPolicy) checkRule(ctx context.Context, ns string, r *api.Rule) error {
	if r.EndpointSelector.Size() > 0 {
		if ok, err := s.checkEPSel(ns, r.EndpointSelector); err != nil {
			return err
		} else if !ok {
			s.AddCode(ctx, 1700, "endpoint")
		}
	}
	for _, ing := range r.Ingress {
		for _, sel := range ing.FromEndpoints {
			if ok, err := s.checkEPSel(ns, sel); err != nil {
				return err
			} else if !ok {
				s.AddCode(ctx, 1700, "ingress")
			}
		}
	}
	for _, eg := range r.Egress {
		for _, sel := range eg.ToEndpoints {
			if ok, err := s.checkEPSel(ns, sel); err != nil {
				return err
			} else if !ok {
				s.AddCode(ctx, 1700, "egress")
			}
		}
	}

	return nil
}

func (s *CiliumNetworkPolicy) checkEPSel(ns string, sel api.EndpointSelector) (bool, error) {
	if sel.Size() == 0 {
		return true, nil
	}

	mm, err := s.matchCEPsBySel(ns, sel)
	if err != nil {
		return false, err
	}

	return len(mm) > 0, nil
}

func (s *CiliumNetworkPolicy) matchCEPsBySel(ns string, sel api.EndpointSelector) ([]string, error) {
	txn := s.db.Txn(false)
	defer txn.Abort()
	txn, it := s.db.MustITForNS(internal.Glossary[cilium.CEP], ns)
	defer txn.Abort()
	mm := make([]string, 0, 10)
	for o := it.Next(); o != nil; o = it.Next() {
		cep, ok := o.(*v2.CiliumEndpoint)
		if !ok {
			return nil, fmt.Errorf("expecting cilium endpoint but got %s", o)
		}
		fqn := client.FQN(cep.Namespace, cep.Name)
		if matchSelector(cep.Labels, sel) {
			mm = append(mm, fqn)
		}
	}

	return mm, nil
}

func matchSelector(labels map[string]string, sel api.EndpointSelector) bool {
	if len(labels) == 0 || sel.Size() == 0 {
		return false
	}
	if matchLabels(labels, sel.MatchLabels) {
		return true
	}

	return matchExp(labels, sel.MatchExpressions)
}

func matchExp(labels map[string]string, ee []slimv1.LabelSelectorRequirement) bool {
	for _, e := range ee {
		if matchSel(labels, e) {
			return true
		}
	}

	return false
}

func matchSel(labels map[string]string, e slimv1.LabelSelectorRequirement) bool {
	key := strings.TrimPrefix(e.Key, "any.")
	_, ok := labels[key]
	if e.Operator == slimv1.LabelSelectorOpDoesNotExist && !ok {
		return true
	}
	if !ok {
		return false
	}

	switch e.Operator {
	case slimv1.LabelSelectorOpNotIn:
		for _, v := range e.Values {
			if v1, ok := labels[key]; ok && v1 == v {
				return false
			}
		}
		return true
	case slimv1.LabelSelectorOpIn:
		for _, v := range e.Values {
			if v == labels[key] {
				return true
			}
		}
		return false
	case slimv1.LabelSelectorOpExists:
		return true
	}

	return false
}

func matchLabels(labels, sel map[string]string) bool {
	var count int
	fmt.Println("LABELS", labels)
	fmt.Println("SEL", sel)
	for k, v := range sel {
		k = strings.TrimPrefix(k, "any.")
		if v1, ok := labels[k]; ok && v == v1 {
			count++
		}
	}

	return count > 0
}
