// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

// Secret tracks Secret sanitization.
type Secret struct {
	*issues.Collector

	db     *db.DB
	system excludedFQN
}

// NewSecret returns a new instance.
func NewSecret(co *issues.Collector, db *db.DB) *Secret {
	return &Secret{
		Collector: co,
		db:        db,
		system: excludedFQN{
			"rx:default-token":                {},
			"rx:^kube-.*/.*-token-":           {},
			"rx:^local-path-storage/.*token-": {},
		},
	}
}

// Lint cleanse the resource.
func (s *Secret) Lint(ctx context.Context) error {
	var refs sync.Map

	if err := cache.NewPod(s.db).PodRefs(&refs); err != nil {
		s.AddErr(ctx, err)
	}
	if err := cache.NewServiceAccount(s.db).ServiceAccountRefs(&refs); err != nil {
		s.AddErr(ctx, err)
	}
	if err := cache.NewIngress(s.db).IngressRefs(&refs); err != nil {
		s.AddErr(ctx, err)
	}
	s.checkStale(ctx, &refs)

	return nil
}

func (s *Secret) checkStale(ctx context.Context, refs *sync.Map) {
	txn, it := s.db.MustITFor(internal.Glossary[internal.SEC])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		sec := o.(*v1.Secret)
		fqn := client.FQN(sec.Namespace, sec.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, sec))

		if s.system.skip(fqn) {
			continue
		}
		refs.Range(func(k, v interface{}) bool {
			return true
		})

		keys, ok := refs.Load(cache.ResFqn(cache.SecretKey, fqn))
		if !ok {
			s.AddCode(ctx, 400)
			continue
		}
		if keys.(internal.StringSet).Has(internal.All) {
			continue
		}

		kk := make(internal.StringSet, len(sec.Data))
		for k := range sec.Data {
			kk.Add(k)
		}
		deltas := keys.(internal.StringSet).Diff(kk)
		for k := range deltas {
			s.AddCode(ctx, 401, k)
		}
	}
}
