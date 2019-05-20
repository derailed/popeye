package cache

import (
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// PodsMetrics represents a Pod metrics cache.
type PodsMetrics struct {
	mx map[string]*mv1beta1.PodMetrics
}

// NewPodsMetrics returns new Pod metrics cache.
func NewPodsMetrics(mx map[string]*mv1beta1.PodMetrics) *PodsMetrics {
	return &PodsMetrics{mx: mx}
}

// ListPodsMetrics returns all available PodMetrics on the cluster.
func (p *PodsMetrics) ListPodsMetrics() map[string]*mv1beta1.PodMetrics {
	return p.mx
}
