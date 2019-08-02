package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/issues"
	pv1beta1 "k8s.io/api/policy/v1beta1"
)

type (
	// PodDisruptionBudget tracks PodDisruptionBudget sanitization.
	PodDisruptionBudget struct {
		*issues.Collector
		PodDisruptionBudgetLister
	}

	// PodDisruptionBudgetLister list available PodDisruptionBudgets on a cluster.
	PodDisruptionBudgetLister interface {
		PodLister
		ListPodDisruptionBudgets() map[string]*pv1beta1.PodDisruptionBudget
	}
)

// NewPodDisruptionBudget returns a new PodDisruptionBudget sanitizer.
func NewPodDisruptionBudget(c *issues.Collector, lister PodDisruptionBudgetLister) *PodDisruptionBudget {
	return &PodDisruptionBudget{
		Collector:                 c,
		PodDisruptionBudgetLister: lister,
	}
}

// Sanitize a configmap.
func (c *PodDisruptionBudget) Sanitize(context.Context) error {
	for fqn, pdb := range c.ListPodDisruptionBudgets() {
		c.InitOutcome(fqn)
		c.checkInUse(fqn, pdb)
	}

	return nil
}

func (c *PodDisruptionBudget) checkInUse(fqn string, pdb *pv1beta1.PodDisruptionBudget) {
	if c.GetPod(pdb.Spec.Selector.MatchLabels) == nil {
		c.AddCode(900, fqn)
		return
	}
}
