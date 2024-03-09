// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"fmt"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type (
	// GatewayClass tracks GatewayClass sanitization.
	GatewayClass struct {
		*issues.Collector

		db *db.DB
	}
)

// NewGatewayClass returns a new instance.
func NewGatewayClass(co *issues.Collector, db *db.DB) *GatewayClass {
	return &GatewayClass{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *GatewayClass) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[internal.GWC])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		gwc := o.(*gwv1.GatewayClass)
		fqn := client.FQN(gwc.Namespace, gwc.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, gwc))
		s.checkRefs(ctx, gwc.Name)
	}

	return nil
}

func (s *GatewayClass) checkRefs(ctx context.Context, n string) {
	txn := s.db.Txn(false)
	defer txn.Abort()
	txn, it := s.db.MustITFor(internal.Glossary[internal.GW])
	defer txn.Abort()

	for o := it.Next(); o != nil; o = it.Next() {
		gw, ok := o.(*gwv1.Gateway)
		if !ok {
			s.AddErr(ctx, fmt.Errorf("expecting gateway but got %T", o))
			continue
		}
		if n == string(gw.Spec.GatewayClassName) {
			return
		}
	}

	s.AddCode(ctx, 400)
}
