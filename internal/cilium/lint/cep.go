// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cilium"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	ilint "github.com/derailed/popeye/internal/lint"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
)

// CiliumEndpoint tracks CiliumEndpoint sanitization.
type CiliumEndpoint struct {
	*issues.Collector
	db *db.DB
}

// NewCiliumEndpoint returns a new instance.
func NewCiliumEndpoint(c *issues.Collector, db *db.DB) *CiliumEndpoint {
	return &CiliumEndpoint{
		Collector: c,
		db:        db,
	}
}

// Lint lints the resource.
func (s *CiliumEndpoint) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[cilium.CEP])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		cep := o.(*v2.CiliumEndpoint)
		fqn := client.FQN(cep.Namespace, cep.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, ilint.SpecFor(fqn, cep))

		if cep.Status.State != "ready" {
			s.AddErr(ctx, fmt.Errorf("cep is not ready"))
		}
		s.checkID(ctx, cep)
		if err := s.checkOwners(ctx, cep); err != nil {
			return err
		}
		if err := s.checkNode(ctx, cep); err != nil {
			return err
		}
	}

	return nil
}

func (s *CiliumEndpoint) checkID(ctx context.Context, cep *v2.CiliumEndpoint) {
	fqn := client.FQN("", strconv.Itoa(int(cep.Status.Identity.ID)))
	_, err := s.db.Find(internal.Glossary[cilium.CID], fqn)
	if err != nil {
		s.AddCode(ctx, 1700, fqn)
	}
}

func (s *CiliumEndpoint) checkOwners(ctx context.Context, cep *v2.CiliumEndpoint) error {
	if len(cep.OwnerReferences) == 0 {
		return errors.New("no owner references found")
	}
	for _, r := range cep.OwnerReferences {
		if config.IsBoolSet(r.Controller) {
			continue
		}
		switch r.Kind {
		case "Pod":
			fqn := client.FQN(cep.Namespace, r.Name)
			o, err := s.db.Find(internal.Glossary[internal.PO], fqn)
			if err != nil {
				s.AddCode(ctx, 1704, fqn)
				continue
			}
			po := o.(*v1.Pod)
			if ph := ilint.Phase(po); ph != "Running" {
				s.AddCode(ctx, 1703, fqn, ph)
			}
		default:
			return fmt.Errorf("nyi - unhandled owner ref kind: %s", r.Kind)
		}
	}

	return nil
}

func (s *CiliumEndpoint) checkNode(ctx context.Context, cep *v2.CiliumEndpoint) error {
	nn, err := s.db.ListNodes()
	if err != nil {
		return err
	}
	nodeIP := cep.Status.Networking.NodeIP
	for _, n := range nn {
		if matchIP(n.Status.Addresses, nodeIP) {
			return nil
		}
	}
	s.AddCode(ctx, 1702, cep.Status.Networking.NodeIP)

	return nil
}

// Helpers...

func matchIP(addrs []v1.NodeAddress, ip string) bool {
	for _, a := range addrs {
		if a.Type != v1.NodeInternalIP {
			continue
		}
		if a.Address == ip {
			return true
		}
	}

	return false
}
