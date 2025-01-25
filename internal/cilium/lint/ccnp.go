// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"fmt"

	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/cilium/cilium/pkg/policy/api"
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cilium"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	ilint "github.com/derailed/popeye/internal/lint"
	v1 "k8s.io/api/core/v1"
)

// CiliumClusterwideNetworkPolicy tracks CiliumClusterwideNetworkPolicy sanitization.
type CiliumClusterwideNetworkPolicy struct {
	*issues.Collector
	db *db.DB
}

// NewCiliumClusterwideNetworkPolicy returns a new instance.
func NewCiliumClusterwideNetworkPolicy(c *issues.Collector, db *db.DB) *CiliumClusterwideNetworkPolicy {
	return &CiliumClusterwideNetworkPolicy{
		Collector: c,
		db:        db,
	}
}

// Lint lints the resource.
func (s *CiliumClusterwideNetworkPolicy) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[cilium.CCNP])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		ccnp := o.(*v2.CiliumClusterwideNetworkPolicy)
		fqn := client.FQN("", ccnp.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, ilint.SpecFor(fqn, ccnp))

		rules := ccnp.Specs
		if ccnp.Spec != nil {
			rules = append(rules, ccnp.Spec)
		}
		for _, r := range rules {
			if err := s.checkRule(ctx, r); err != nil {
				s.AddErr(ctx, err)
			}
		}
	}

	return nil
}

func (s *CiliumClusterwideNetworkPolicy) checkRule(ctx context.Context, r *api.Rule) error {
	if ok, err := s.checkEPSel(r.EndpointSelector); err != nil {
		return err
	} else if !ok {
		s.AddCode(ctx, 1700, "endpoint")
	}

	if ok, err := s.checkNodeSel(r.NodeSelector); err != nil {
		return err
	} else if !ok {
		s.AddCode(ctx, 1701)
	}

	for _, ing := range r.Ingress {
		for _, sel := range ing.FromEndpoints {
			if ok, err := s.checkEPSel(sel); err != nil {
				return err
			} else if !ok {
				s.AddCode(ctx, 1700, "ingress")
			}
		}
	}
	for _, eg := range r.Egress {
		for _, sel := range eg.ToEndpoints {
			if ok, err := s.checkEPSel(sel); err != nil {
				return err
			} else if !ok {
				s.AddCode(ctx, 1700, "egress")
			}
		}
	}

	return nil
}

func (s *CiliumClusterwideNetworkPolicy) checkEPSel(sel api.EndpointSelector) (bool, error) {
	if sel.Size() == 0 {
		return true, nil
	}

	mm, err := s.matchCEPsBySel(sel)
	if err != nil {
		return false, err
	}

	return len(mm) > 0, nil
}

func (s *CiliumClusterwideNetworkPolicy) checkNodeSel(sel api.EndpointSelector) (bool, error) {
	if sel.Size() == 0 {
		return true, nil
	}

	mm, err := s.matchNodesBySel(sel)
	if err != nil {
		return false, err
	}

	return len(mm) > 0, nil
}

func (s *CiliumClusterwideNetworkPolicy) matchNodesBySel(sel api.EndpointSelector) ([]string, error) {
	txn := s.db.Txn(false)
	defer txn.Abort()
	txn, it := s.db.MustITFor(internal.Glossary[internal.NO])
	defer txn.Abort()
	mm := make([]string, 0, 10)
	for o := it.Next(); o != nil; o = it.Next() {
		no, ok := o.(*v1.Node)
		if !ok {
			return nil, fmt.Errorf("expecting node but got %s", o)
		}
		ll := make([]string, 0, len(no.Labels))
		for k, v := range no.Labels {
			ll = append(ll, k+"="+v)
		}
		if matchSelector(client.AllNamespaces, ll, sel) {
			mm = append(mm, client.FQN("", no.Name))
		}
	}

	return mm, nil
}

func (s *CiliumClusterwideNetworkPolicy) matchCEPsBySel(sel api.EndpointSelector) ([]string, error) {
	txn := s.db.Txn(false)
	defer txn.Abort()
	txn, it := s.db.MustITFor(internal.Glossary[cilium.CEP])
	defer txn.Abort()
	mm := make([]string, 0, 10)
	for o := it.Next(); o != nil; o = it.Next() {
		cep, ok := o.(*v2.CiliumEndpoint)
		if !ok {
			return nil, fmt.Errorf("expecting cilium endpoint but got %s", o)
		}
		if cep.Status.Identity == nil {
			continue
		}
		if matchSelector(cep.Namespace, cep.Status.Identity.Labels, sel) {
			mm = append(mm, client.FQN(cep.Namespace, cep.Name))
		}
	}

	return mm, nil
}
