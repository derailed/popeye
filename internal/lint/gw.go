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
	"github.com/rs/zerolog/log"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type (
	// Gateway tracks Gateway sanitization.
	Gateway struct {
		*issues.Collector

		db *db.DB
	}
)

// NewGateway returns a new instance.
func NewGateway(co *issues.Collector, db *db.DB) *Gateway {
	return &Gateway{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *Gateway) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[internal.GW])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		gw := o.(*gwv1.Gateway)
		fqn := client.FQN(gw.Namespace, gw.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, gw))
		s.checkRefs(ctx, gw)
	}

	return nil
}

func (s *Gateway) checkRefs(ctx context.Context, gw *gwv1.Gateway) {
	txn, it, err := s.db.ITFor(internal.Glossary[internal.GWC])
	if err != nil {
		log.Warn().Err(err).Msg("no gateway class located. Skipping gw ref check")
		return
	}
	defer txn.Abort()

	for o := it.Next(); o != nil; o = it.Next() {
		gwc, ok := o.(*gwv1.GatewayClass)
		if !ok {
			s.AddErr(ctx, fmt.Errorf("expecting gatewayclass but got %T", o))
			continue
		}
		if gwc.Name == string(gw.Spec.GatewayClassName) {
			return
		}
	}

	s.AddCode(ctx, 407, gw.Kind, "GatewayClass", gw.Spec.GatewayClassName)
}
