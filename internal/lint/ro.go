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
	rbacv1 "k8s.io/api/rbac/v1"
)

type (
	// Role tracks Role sanitization.
	Role struct {
		*issues.Collector

		db *db.DB
	}
)

// NewRole returns a new instance.
func NewRole(c *issues.Collector, db *db.DB) *Role {
	return &Role{
		Collector: c,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *Role) Lint(ctx context.Context) error {
	var refs sync.Map

	crb := cache.NewClusterRoleBinding(s.db)
	crb.ClusterRoleRefs(&refs)

	rb := cache.NewRoleBinding(s.db)
	rb.RoleRefs(&refs)

	s.checkInUse(ctx, &refs)

	return nil
}

func (s *Role) checkInUse(ctx context.Context, refs *sync.Map) {
	txn, it := s.db.MustITFor(internal.Glossary[internal.RO])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		ro := o.(*rbacv1.Role)
		fqn := client.FQN(ro.Namespace, ro.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, ro))

		if _, ok := refs.Load(cache.ResFqn(cache.RoleKey, fqn)); !ok {
			s.AddCode(ctx, 400)
		}
	}
}
