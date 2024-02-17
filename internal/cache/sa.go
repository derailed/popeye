// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"errors"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	v1 "k8s.io/api/core/v1"
)

// ServiceAccount tracks serviceaccounts.
type ServiceAccount struct {
	db *db.DB
}

// NewServiceAccount returns a new serviceaccount loader.
func NewServiceAccount(db *db.DB) *ServiceAccount {
	return &ServiceAccount{db: db}
}

// ServiceAccountRefs computes all serviceaccount external references.
func (s *ServiceAccount) ServiceAccountRefs(refs *sync.Map) error {
	txn, it := s.db.MustITFor(internal.Glossary[internal.SA])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		sa, ok := o.(*v1.ServiceAccount)
		if !ok {
			return errors.New("expected sa")
		}
		namespaceRefs(sa.Namespace, refs)
		for _, s := range sa.Secrets {
			key := ResFqn(SecretKey, FQN(s.Namespace, s.Name))
			refs.Store(key, internal.AllKeys)
		}
		for _, s := range sa.ImagePullSecrets {
			key := ResFqn(SecretKey, FQN(sa.Namespace, s.Name))
			refs.Store(key, internal.AllKeys)
		}

	}

	return nil
}
