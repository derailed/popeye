// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"fmt"
	"strconv"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type (
	// HTTPRoute tracks HTTPRoute sanitization.
	HTTPRoute struct {
		*issues.Collector

		db *db.DB
	}
)

// NewHTTPRoute returns a new instance.
func NewHTTPRoute(co *issues.Collector, db *db.DB) *HTTPRoute {
	return &HTTPRoute{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *HTTPRoute) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[internal.GWR])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		gwr := o.(*gwv1.HTTPRoute)
		fqn := client.FQN(gwr.Namespace, gwr.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, gwr))
		s.checkRoute(ctx, fqn, gwr)
	}

	return nil
}

// Check service ref
func (s *HTTPRoute) checkRoute(ctx context.Context, fqn string, gwr *gwv1.HTTPRoute) {
	for _, r := range gwr.Spec.ParentRefs {
		switch {
		case r.Group == nil:
			var kind string
			if r.Kind == nil {
				kind = "Gateway"
			} else {
				kind = string(*r.Kind)
			}
			switch kind {
			case "Gateway":
				s.checkGWRef(ctx, gwr.Namespace, &r)
			case "Service":
				s.checkSvcRef(ctx, gwr.Namespace, &r)
			default:
				s.AddErr(ctx, fmt.Errorf("unhandled parent kind: %s", kind))
			}
		case *r.Group == "", *r.Group == "Service":
			s.checkSvcRef(ctx, gwr.Namespace, &r)
		}
	}

	for _, r := range gwr.Spec.Rules {
		for _, be := range r.BackendRefs {
			switch {
			case be.Kind == nil, *be.Kind == "Service":
				s.checkSvcBE(ctx, gwr.Namespace, &be.BackendRef)
			}
		}
	}
}

func (s *HTTPRoute) checkSvcBE(ctx context.Context, ns string, be *gwv1.BackendRef) {
	if be.BackendObjectReference.Kind == nil || *be.BackendObjectReference.Kind == "Service" {
		txn := s.db.Txn(false)
		defer txn.Abort()

		if be.Namespace != nil {
			ns = string(*be.Namespace)
		}
		fqn := client.FQN(ns, string(be.Name))
		o, err := s.db.Find(internal.Glossary[internal.SVC], fqn)
		if err != nil {
			s.AddCode(ctx, 407, "Route", "Service", fqn)
			return
		}
		svc, ok := o.(*v1.Service)
		if !ok {
			s.AddErr(ctx, fmt.Errorf("expecting service but got %T", o))
			return
		}
		if be.Port == nil {
			return
		}
		for _, p := range svc.Spec.Ports {
			if p.Port == int32(*be.Port) {
				return
			}
		}
		s.AddCode(ctx, 1106, strconv.Itoa(int(*be.Port)))
	}
}

func (s *HTTPRoute) checkGWRef(ctx context.Context, ns string, ref *gwv1.ParentReference) {
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}
	fqn := client.FQN(ns, string(ref.Name))
	_, err := s.db.Find(internal.Glossary[internal.GW], fqn)
	if err != nil {
		s.AddCode(ctx, 407, "HTTPRoute", "Gateway", fqn)
	}
}

func (s *HTTPRoute) checkSvcRef(ctx context.Context, ns string, ref *gwv1.ParentReference) {
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}
	fqn := client.FQN(ns, string(ref.Name))
	_, err := s.db.Find(internal.Glossary[internal.SVC], fqn)
	if err != nil {
		s.AddCode(ctx, 407, "HTTPRoute", "Service", fqn)
	}
}
