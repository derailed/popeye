// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
)

type (
	// Ingress tracks Ingress sanitization.
	Ingress struct {
		*issues.Collector

		db *db.DB
	}
)

// NewIngress returns a new instance.
func NewIngress(co *issues.Collector, db *db.DB) *Ingress {
	return &Ingress{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *Ingress) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[internal.ING])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		ing := o.(*netv1.Ingress)
		fqn := client.FQN(ing.Namespace, ing.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, ing))

		for _, ing := range ing.Status.LoadBalancer.Ingress {
			for _, p := range ing.Ports {
				if p.Error != nil {
					s.AddCode(ctx, 1400, *p.Error)
				}
			}
		}
		for _, r := range ing.Spec.Rules {
			http := r.IngressRuleValue.HTTP
			if http == nil {
				continue
			}
			for _, h := range http.Paths {
				s.checkBackendSvc(ctx, ing.Namespace, h.Backend.Service)
				s.checkBackendRef(ctx, ing.Namespace, h.Backend.Resource)
			}
		}
	}

	return nil
}

func (s *Ingress) checkBackendRef(ctx context.Context, ns string, be *v1.TypedLocalObjectReference) {
	if be == nil {
		return
	}
	s.AddErr(ctx, errors.New("Ingress local obj refs not supported"))
}

func (s *Ingress) checkBackendSvc(ctx context.Context, ns string, be *netv1.IngressServiceBackend) {
	if be == nil {
		return
	}
	o, err := s.db.Find(internal.Glossary[internal.SVC], cache.FQN(ns, be.Name))
	if err != nil {
		s.AddCode(ctx, 1401, be.Name)
		return
	}
	isvc, ok := o.(*v1.Service)
	if !ok {
		s.AddErr(ctx, fmt.Errorf("expecting service but got %T", o))
		return
	}
	if !s.findPortByNumberOrName(ctx, isvc.Spec.Ports, be.Port) {
		s.AddCode(ctx, 1402, fmt.Sprintf("%s:%d", be.Port.Name, be.Port.Number))
	}
	if be.Port.Name == "" {
		if be.Port.Number == 0 {
			s.AddCode(ctx, 1404)
			return
		}
		s.AddCode(ctx, 1403, be.Port.Number)
	}
}

func (s *Ingress) findPortByNumberOrName(ctx context.Context, pp []v1.ServicePort, port netv1.ServiceBackendPort) bool {
	for _, p := range pp {
		if p.Name == port.Name {
			return true
		}
		if p.Port == port.Number {
			return true
		}
	}

	return false
}
