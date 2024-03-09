// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	polv1 "k8s.io/api/policy/v1"
)

// PodDisruptionBudget tracks PodDisruptionBudget sanitization.
type PodDisruptionBudget struct {
	*issues.Collector

	db *db.DB
}

// NewPodDisruptionBudget returns a new PodDisruptionBudget linter.
func NewPodDisruptionBudget(c *issues.Collector, db *db.DB) *PodDisruptionBudget {
	return &PodDisruptionBudget{
		Collector: c,
		db:        db,
	}
}

// Lint cleanse the resource.
func (p *PodDisruptionBudget) Lint(ctx context.Context) error {
	txn, it := p.db.MustITFor(internal.Glossary[internal.PDB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		pdb := o.(*polv1.PodDisruptionBudget)
		fqn := client.FQN(pdb.Namespace, pdb.Name)
		p.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, pdb))

		p.checkInUse(ctx, pdb)
	}

	return nil
}

func (p *PodDisruptionBudget) checkInUse(ctx context.Context, pdb *polv1.PodDisruptionBudget) {
	pp, err := p.db.FindPodsBySel(pdb.Namespace, pdb.Spec.Selector)
	if err != nil || len(pp) == 0 {
		p.AddCode(ctx, 900, dumpSel(pdb.Spec.Selector))
		return
	}
}
