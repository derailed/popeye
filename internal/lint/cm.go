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

// ConfigMap tracks ConfigMap sanitization.
type ConfigMap struct {
	*issues.Collector
	db     *db.DB
	system excludedFQN
}

// NewConfigMap returns a new instance.
func NewConfigMap(c *issues.Collector, db *db.DB) *ConfigMap {
	return &ConfigMap{
		Collector: c,
		db:        db,
		system: excludedFQN{
			"rx:^kube-public":     {},
			"rx:kube-root-ca.crt": {},
		},
	}
}

// Lint lints the resource.
func (s *ConfigMap) Lint(ctx context.Context) error {
	var cmRefs sync.Map
	if err := cache.NewPod(s.db).PodRefs(&cmRefs); err != nil {
		return err
	}

	return s.checkStale(ctx, &cmRefs)
}

func (s *ConfigMap) checkStale(ctx context.Context, refs *sync.Map) error {
	txn, it := s.db.MustITFor(internal.Glossary[internal.CM])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		cm := o.(*v1.ConfigMap)
		fqn := client.FQN(cm.Namespace, cm.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, cm))
		if s.system.skip(fqn) {
			continue
		}

		keys, ok := refs.Load(cache.ResFqn(cache.ConfigMapKey, fqn))
		if !ok {
			s.AddCode(ctx, 400)
			continue
		}
		if keys.(internal.StringSet).Has(internal.All) {
			continue
		}
		kk := make(internal.StringSet, len(cm.Data))
		for k := range cm.Data {
			kk.Add(k)
		}
		deltas := keys.(internal.StringSet).Diff(kk)
		for k := range deltas {
			s.AddCode(ctx, 401, k)
		}
	}

	return nil
}
