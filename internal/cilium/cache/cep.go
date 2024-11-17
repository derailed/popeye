// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"fmt"
	"strconv"
	"sync"

	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/derailed/popeye/internal"
	icache "github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/cilium"
	"github.com/derailed/popeye/internal/db"
)

const CIDKey = "cid"

// CiliumEndpoint represents a CiliumEndpoint cache.
type CiliumEndpoint struct {
	db *db.DB
}

// NewCiliumEndpoint returns a CiliumEndpoint cache.
func NewCiliumEndpoint(dba *db.DB) *CiliumEndpoint {
	return &CiliumEndpoint{db: dba}
}

// CEPRefs computes all CiliumEndpoints external references.
func (p *CiliumEndpoint) CEPRefs(refs *sync.Map) error {
	txn, it := p.db.MustITFor(internal.Glossary[cilium.CEP])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		cep, ok := o.(*v2.CiliumEndpoint)
		if !ok {
			return fmt.Errorf("expected a CiliumEndpoint but got %T", o)
		}
		if cep.Status.Identity != nil {
			key := icache.ResFqn(CIDKey, icache.FQN("", strconv.Itoa(int(cep.Status.Identity.ID))))
			refs.Store(key, internal.AllKeys)
		}
	}

	return nil
}
