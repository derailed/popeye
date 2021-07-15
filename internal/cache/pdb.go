package cache

import (
	polv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodDisruptionBudgetKey tracks PodDisruptionBudget resource references
const PodDisruptionBudgetKey = "pdb"

// PodDisruptionBudget represents PodDisruptionBudget cache.
type PodDisruptionBudget struct {
	cms map[string]*polv1beta1.PodDisruptionBudget
}

// NewPodDisruptionBudget returns a new PodDisruptionBudget cache.
func NewPodDisruptionBudget(cms map[string]*polv1beta1.PodDisruptionBudget) *PodDisruptionBudget {
	return &PodDisruptionBudget{cms: cms}
}

// ListPodDisruptionBudgets returns all available PodDisruptionBudgets on the cluster.
func (c *PodDisruptionBudget) ListPodDisruptionBudgets() map[string]*polv1beta1.PodDisruptionBudget {
	return c.cms
}

// ForLabels returns a pdb whose selector match the given labels. Returns nil if no match.
func (c *PodDisruptionBudget) ForLabels(labels map[string]string) *polv1beta1.PodDisruptionBudget {
	for _, pdb := range c.ListPodDisruptionBudgets() {
		m, err := metav1.LabelSelectorAsMap(pdb.Spec.Selector)
		if err != nil {
			continue
		}
		if matchLabels(labels, m) {
			return pdb
		}
	}
	return nil
}
