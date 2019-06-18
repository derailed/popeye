package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/issues"
	pv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
		c.AddWarnf(fqn, "Used? No pods match selector")
		return
	}

	min := pdb.Spec.MinAvailable
	if min != nil && min.Type == intstr.Int && min.IntValue() > int(pdb.Status.CurrentHealthy) {
		c.AddWarnf(fqn, "MinAvailable (%d) is greater than the number of pods(%d) currently running", min.IntValue(), pdb.Status.CurrentHealthy)
	}
}
