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
	"github.com/derailed/popeye/internal/rules"
	rbacv1 "k8s.io/api/rbac/v1"
)

type excludedFQN map[rules.Expression]struct{}

func (e excludedFQN) skip(fqn string) bool {
	if _, ok := e[rules.Expression(fqn)]; ok {
		return true
	}
	for k := range e {
		if k.IsRX() && k.MatchRX(fqn) {
			return true
		}
	}

	return false
}

// ClusterRole tracks ClusterRole sanitization.
type ClusterRole struct {
	*issues.Collector

	db     *db.DB
	system excludedFQN
}

// NewClusterRole returns a new instance.
func NewClusterRole(c *issues.Collector, db *db.DB) *ClusterRole {
	return &ClusterRole{
		Collector: c,
		db:        db,
		system: excludedFQN{
			"admin":       {},
			"edit":        {},
			"view":        {},
			"rx:^system:": {},
		},
	}
}

// Lint sanitizes the resource.
func (s *ClusterRole) Lint(ctx context.Context) error {
	var crRefs, agRefs sync.Map
	crb := cache.NewClusterRoleBinding(s.db)
	crb.ClusterRoleRefs(&crRefs)
	rb := cache.NewRoleBinding(s.db)
	rb.RoleRefs(&crRefs)
	cr := cache.NewClusterRole(s.db)
	cr.AggregationMatchers(&agRefs)
	s.checkStale(ctx, &crRefs, &agRefs)

	return nil
}

func (s *ClusterRole) checkStale(ctx context.Context, refs *sync.Map, agRefs *sync.Map) {
	txn, it := s.db.MustITFor(internal.Glossary[internal.CR])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		cr := o.(*rbacv1.ClusterRole)
		fqn := client.FQN(cr.Namespace, cr.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, cr))
		if s.system.skip(fqn) {
			continue
		}
		partialRole := false
		for key, value := range cr.Labels {
			expectedValue, ok := agRefs.Load(key)
			if ok && value == expectedValue {
				partialRole = true
			}
		}
		if partialRole {
			continue
		}
		if _, ok := refs.Load(cache.ResFqn(cache.ClusterRoleKey, fqn)); !ok {
			s.AddCode(ctx, 400)
		}
	}
}
