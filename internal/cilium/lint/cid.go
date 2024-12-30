// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"fmt"
	"sync"

	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/derailed/popeye/internal"
	icache "github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/cilium"
	"github.com/derailed/popeye/internal/cilium/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	ilint "github.com/derailed/popeye/internal/lint"
)

// CiliumIdentity tracks CiliumIdentity sanitization.
type CiliumIdentity struct {
	*issues.Collector
	db *db.DB
}

// NewCiliumIdentity returns a new instance.
func NewCiliumIdentity(c *issues.Collector, db *db.DB) *CiliumIdentity {
	return &CiliumIdentity{
		Collector: c,
		db:        db,
	}
}

// Lint lints the resource.
func (s *CiliumIdentity) Lint(ctx context.Context) error {
	var refs sync.Map
	if err := cache.NewCiliumEndpoint(s.db).CEPRefs(&refs); err != nil {
		return err
	}

	txn, it := s.db.MustITFor(internal.Glossary[cilium.CID])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		cid := o.(*v2.CiliumIdentity)
		fqn := client.FQN(cid.Namespace, cid.Name)
		id := client.FQN(extractNS(cid), cid.Name)
		s.InitOutcome(id)
		ctx = internal.WithSpec(ctx, ilint.SpecFor(id, cid))
		if err := s.checkStale(ctx, fqn, &refs); err != nil {
			return err
		}
		s.checkNS(ctx, cid)
		if err := s.checkSA(ctx, cid); err != nil {
			return err
		}
	}

	return nil
}

func (s *CiliumIdentity) checkStale(ctx context.Context, fqn string, refs *sync.Map) error {
	if _, ok := refs.Load(icache.ResFqn(cache.CIDKey, fqn)); !ok {
		s.AddCode(ctx, 1600)
	}

	return nil
}

const (
	k8sNSLabel     = "io.kubernetes.pod.namespace"
	k8sSecNSLabel  = "k8s:io.kubernetes.pod.namespace"
	k8sSecNS1Label = "k8s:io.cilium.k8s.namespace.labels.kubernetes.io/metadata.name"
	k8sSALabel     = "io.cilium.k8s.policy.serviceaccount"
)

func extractNS(cid *v2.CiliumIdentity) string {
	if ns, ok := cid.Labels[k8sNSLabel]; ok {
		return ns
	}
	if ns, ok := cid.SecurityLabels[k8sSecNSLabel]; ok {
		return ns
	}

	return client.BlankNamespace
}

func (s *CiliumIdentity) checkNS(ctx context.Context, cid *v2.CiliumIdentity) {
	ns, ok := cid.Labels[k8sNSLabel]
	if !ok {
		s.AddCode(ctx, 1601, k8sNSLabel)
	}
	_, err := s.db.Find(internal.Glossary[internal.NS], ns)
	if err != nil {
		s.AddCode(ctx, 1602, ns)
		return
	}
	sns, ok := cid.SecurityLabels[k8sSecNSLabel]
	if !ok {
		s.AddCode(ctx, 1603, k8sSecNSLabel)
		return
	}
	if ns != sns {
		s.AddCode(ctx, 1604, ns, sns)
	}
}

func (s *CiliumIdentity) checkSA(ctx context.Context, cid *v2.CiliumIdentity) error {
	ns, ok := cid.Labels[k8sNSLabel]
	if !ok {
		return fmt.Errorf("unable to locate cid namespace")
	}
	sa, ok := cid.Labels[k8sSALabel]
	if !ok {
		return fmt.Errorf("unable to locate cid serviceaccount")
	}
	txn := s.db.Txn(false)
	defer txn.Abort()
	saFQN := icache.FQN(ns, sa)
	o, err := txn.First(internal.Glossary[internal.SA].String(), "id", saFQN)
	if err != nil || o == nil {
		s.AddCode(ctx, 307, "CiliumIdentity", saFQN)
		return nil
	}

	return nil
}
