package cache

import (
	autoscalingv1 "k8s.io/api/autoscaling/v1"
)

// HorizontalPodAutoscaler represents a collection of HorizontalPodAutoScalers available on a cluster.
type HorizontalPodAutoscaler struct {
	hpas map[string]*autoscalingv1.HorizontalPodAutoscaler
}

// NewHorizontalPodAutoscaler returns a new HorizontalPodAutoScaler.
func NewHorizontalPodAutoscaler(svcs map[string]*autoscalingv1.HorizontalPodAutoscaler) *HorizontalPodAutoscaler {
	return &HorizontalPodAutoscaler{svcs}
}

// ListHorizontalPodAutoscalers returns all available HorizontalPodAutoScalers on the cluster.
func (h *HorizontalPodAutoscaler) ListHorizontalPodAutoscalers() map[string]*autoscalingv1.HorizontalPodAutoscaler {
	return h.hpas
}
