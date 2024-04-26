// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	rbacv1 "k8s.io/api/rbac/v1"
)

// ClusterRole represents ClusterRole cache.
type ClusterRole struct {
	db *db.DB
}

// NewClusterRole returns a new ClusterRole cache.
func NewClusterRole(db *db.DB) *ClusterRole {
	return &ClusterRole{db: db}
}

// RoleRefs computes all role external references.
func (r *ClusterRole) AggregationMatchers(refs *sync.Map) {
	txn, it := r.db.MustITFor(internal.Glossary[internal.CR])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		cr := o.(*rbacv1.ClusterRole)
		if cr.AggregationRule != nil {
			for _, lbs := range cr.AggregationRule.ClusterRoleSelectors {
				for k, v := range lbs.MatchLabels {
					refs.Store(k, v)
				}
			}
		}
	}
}
