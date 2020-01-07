package cache

import (
	v1beta1 "k8s.io/api/policy/v1beta1"
)

// PodDisruptionBudgetKey tracks PodDisruptionBudget resource references
const PodDisruptionBudgetKey = "pdb"

// PodDisruptionBudget represents PodDisruptionBudget cache.
type PodDisruptionBudget struct {
	cms map[string]*v1beta1.PodDisruptionBudget
}

// NewPodDisruptionBudget returns a new PodDisruptionBudget cache.
func NewPodDisruptionBudget(cms map[string]*v1beta1.PodDisruptionBudget) *PodDisruptionBudget {
	return &PodDisruptionBudget{cms: cms}
}

// ListPodDisruptionBudgets returns all available PodDisruptionBudgets on the cluster.
func (c *PodDisruptionBudget) ListPodDisruptionBudgets() map[string]*v1beta1.PodDisruptionBudget {
	return c.cms
}

// ForLabels returns a pdb whose selector match the given labels. Returns nil if no match.
func (c *PodDisruptionBudget) ForLabels(labels map[string]string) *v1beta1.PodDisruptionBudget {
	for _, pdb := range c.ListPodDisruptionBudgets() {
		if matchLabels(labels, pdb.Spec.Selector.MatchLabels) {
			return pdb
		}
	}
	return nil
}
