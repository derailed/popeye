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
	// RoleBinding tracks RoleBinding sanitization.
	RoleBinding struct {
		*issues.Collector

		db *db.DB
	}
)

// NewRoleBinding returns a new instance.
func NewRoleBinding(c *issues.Collector, db *db.DB) *RoleBinding {
	return &RoleBinding{
		Collector: c,
		db:        db,
	}
}

// Lint cleanse the resource..
func (r *RoleBinding) Lint(ctx context.Context) error {
	r.checkInUse(ctx)

	return nil
}

func (r *RoleBinding) checkInUse(ctx context.Context) {
	txn, it := r.db.MustITFor(internal.Glossary[internal.ROB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		rb := o.(*rbacv1.RoleBinding)
		fqn := client.FQN(rb.Namespace, rb.Name)
		r.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, rb))

		switch rb.RoleRef.Kind {
		case "ClusterRole":
			if !r.db.Exists(internal.Glossary[internal.CR], rb.RoleRef.Name) {
				r.AddCode(ctx, 1300, rb.RoleRef.Kind, rb.RoleRef.Name)
			}
		case "Role":
			rFQN := cache.FQN(rb.Namespace, rb.RoleRef.Name)
			if !r.db.Exists(internal.Glossary[internal.RO], rFQN) {
				r.AddCode(ctx, 1300, rb.RoleRef.Kind, rFQN)
			}
		}
	}
}

func boundDefaultSA(db *db.DB) bool {
	txn, it := db.MustITFor(internal.Glossary[internal.ROB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		rb := o.(*rbacv1.RoleBinding)
		if rb.Namespace != client.DefaultNamespace || rb.RoleRef.Kind == "ClusterRole" {
			continue
		}
		if rb.RoleRef.APIGroup == "" && rb.RoleRef.Kind == "ServiceAccount" && rb.RoleRef.Name == "default" {
			return true
		}
	}

	txn, it = db.MustITFor(internal.Glossary[internal.CRB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		rb := o.(*rbacv1.ClusterRoleBinding)
		if rb.RoleRef.Kind == "ClusterRole" {
			continue
		}
		if rb.RoleRef.APIGroup == "" && rb.RoleRef.Kind == "ServiceAccount" && rb.RoleRef.Name == "default" {
			return true
		}
	}

	return false
}
