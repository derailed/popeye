// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	rbacv1 "k8s.io/api/rbac/v1"
)

type (
	// ClusterRoleBinding tracks ClusterRoleBinding sanitization.
	ClusterRoleBinding struct {
		*issues.Collector

		db *db.DB
	}
)

// NewClusterRoleBinding returns a new instance.
func NewClusterRoleBinding(c *issues.Collector, db *db.DB) *ClusterRoleBinding {
	return &ClusterRoleBinding{
		Collector: c,
		db:        db,
	}
}

// Lint sanitizes the resource.
func (c *ClusterRoleBinding) Lint(ctx context.Context) error {
	c.checkInUse(ctx)

	return nil
}

func (c *ClusterRoleBinding) checkInUse(ctx context.Context) {
	txn, it := c.db.MustITFor(internal.Glossary[internal.CRB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		crb := o.(*rbacv1.ClusterRoleBinding)
		fqn := client.FQN(crb.Namespace, crb.Name)

		c.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, crb))

		switch crb.RoleRef.Kind {
		case "ClusterRole":
			if !c.db.Exists(internal.Glossary[internal.CR], crb.RoleRef.Name) {
				c.AddCode(ctx, 1300, crb.RoleRef.Kind, crb.RoleRef.Name)
			}
		case "Role":
			rFQN := cache.FQN(crb.Namespace, crb.RoleRef.Name)
			if !c.db.Exists(internal.Glossary[internal.RO], rFQN) {
				c.AddCode(ctx, 1300, crb.RoleRef.Kind, rFQN)
			}
		}
		for _, s := range crb.Subjects {
			if s.Kind == "ServiceAccount" {
				safqn := cache.FQN(s.Namespace, s.Name)
				if !c.db.Exists(internal.Glossary[internal.SA], safqn) {
					c.AddCode(ctx, 1300, s.Kind, safqn)
				}
			}
		}
	}
}
