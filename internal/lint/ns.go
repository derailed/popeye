// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"errors"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

// Namespace represents a Namespace linter.
type Namespace struct {
	*issues.Collector

	db *db.DB
}

// NewNamespace returns a new instance.
func NewNamespace(co *issues.Collector, db *db.DB) *Namespace {
	return &Namespace{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *Namespace) Lint(ctx context.Context) error {
	used := make(map[string]struct{})
	if err := s.ReferencedNamespaces(used); err != nil {
		s.AddErr(ctx, err)
	}

	cns, ok := ctx.Value(internal.KeyNamespaceName).(string)
	if !ok {
		cns = client.AllNamespaces
	}

	txn, it := s.db.MustITFor(internal.Glossary[internal.NS])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		ns, ok := o.(*v1.Namespace)
		if !ok {
			return errors.New("expected ns")
		}
		fqn := ns.Name
		if client.IsNamespaced(cns) && fqn != cns {
			continue
		}
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, ns))

		if s.checkActive(ctx, ns.Status.Phase) {
			if _, ok := used[fqn]; !ok {
				s.AddCode(ctx, 400)
			}
		}
	}

	return nil
}

// ReferencedNamespaces fetch all namespaces referenced by pods and service accounts.
func (s *Namespace) ReferencedNamespaces(res map[string]struct{}) error {
	var refs sync.Map
	pod := cache.NewPod(s.db)
	if err := pod.PodRefs(&refs); err != nil {
		return err
	}
	sa := cache.NewServiceAccount(s.db)
	if err := sa.ServiceAccountRefs(&refs); err != nil {
		return err
	}
	if ss, ok := refs.Load("ns"); ok {
		for ns := range ss.(internal.StringSet) {
			res[ns] = struct{}{}
		}
	}

	return nil
}

func (s *Namespace) checkActive(ctx context.Context, p v1.NamespacePhase) bool {
	if !isNSActive(p) {
		s.AddCode(ctx, 800)
		return false
	}

	return true
}

// ----------------------------------------------------------------------------
// Helpers...

func isNSActive(phase v1.NamespacePhase) bool {
	return phase == v1.NamespaceActive
}
