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

// ClusterRoleBinding represents ClusterRoleBinding cache.
type ClusterRoleBinding struct {
	db *db.DB
}

// NewClusterRoleBinding returns a new ClusterRoleBinding cache.
func NewClusterRoleBinding(db *db.DB) *ClusterRoleBinding {
	return &ClusterRoleBinding{db: db}
}

// ClusterRoleRefs computes all clusterrole external references.
func (c *ClusterRoleBinding) ClusterRoleRefs(refs *sync.Map) {
	txn, it := c.db.MustITFor(internal.Glossary[internal.CRB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		crb := o.(*rbacv1.ClusterRoleBinding)
		fqn := client.FQN(crb.Namespace, crb.Name)
		key := ResFqn(strings.ToLower(crb.RoleRef.Kind), FQN(crb.Namespace, crb.RoleRef.Name))
		if c, ok := refs.Load(key); ok {
			c.(internal.StringSet).Add(fqn)
		} else {
			refs.Store(key, internal.StringSet{fqn: internal.Blank})
		}
	}
}
