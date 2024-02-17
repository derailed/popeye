// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"strings"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	rbacv1 "k8s.io/api/rbac/v1"
)

// RoleKey represents a role identifier.
const RoleKey = "role"

// RoleBinding represents RoleBinding cache.
type RoleBinding struct {
	db *db.DB
}

// NewRoleBinding returns a new RoleBinding cache.
func NewRoleBinding(db *db.DB) *RoleBinding {
	return &RoleBinding{db: db}
}

// RoleRefs computes all role external references.
func (r *RoleBinding) RoleRefs(refs *sync.Map) {
	txn, it := r.db.MustITFor(internal.Glossary[internal.ROB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		rb := o.(*rbacv1.RoleBinding)
		fqn := client.FQN(rb.Namespace, rb.Name)
		cfqn := FQN(rb.Namespace, rb.RoleRef.Name)
		if rb.RoleRef.Kind == "ClusterRole" {
			cfqn = client.FQN("", rb.RoleRef.Name)
		}
		key := ResFqn(strings.ToLower(rb.RoleRef.Kind), cfqn)
		if c, ok := refs.LoadOrStore(key, internal.StringSet{fqn: internal.Blank}); ok {
			c.(internal.StringSet).Add(fqn)
		}
	}
}
