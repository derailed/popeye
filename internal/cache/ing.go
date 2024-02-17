// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"errors"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	netv1 "k8s.io/api/networking/v1"
)

// IngressKey tracks Ingress resource references
const IngressKey = "ing"

// Ingress represents Ingress cache.
type Ingress struct {
	db *db.DB
}

// NewIngress returns a new Ingress cache.
func NewIngress(db *db.DB) *Ingress {
	return &Ingress{db: db}
}

// IngressRefs computes all ingress external references.
func (d *Ingress) IngressRefs(refs *sync.Map) error {
	txn, it := d.db.MustITFor(internal.Glossary[internal.ING])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		ing, ok := o.(*netv1.Ingress)
		if !ok {
			return errors.New("expected ing")
		}
		for _, tls := range ing.Spec.TLS {
			d.trackReference(refs, ResFqn(SecretKey, FQN(ing.Namespace, tls.SecretName)))
		}
	}

	return nil
}

func (d *Ingress) trackReference(refs *sync.Map, key string) {
	refs.Store(key, internal.AllKeys)
}
