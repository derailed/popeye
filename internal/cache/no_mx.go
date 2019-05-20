package cache

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// NodesMetrics represents a Node metrics cache.
type NodesMetrics struct {
	mx map[string]*mv1beta1.NodeMetrics
}

// NewNodesMetrics returns new Node metrics cache.
func NewNodesMetrics(mx map[string]*mv1beta1.NodeMetrics) *NodesMetrics {
	return &NodesMetrics{mx: mx}
}

// ListNodesMetrics returns all available NodeMetrics on the cluster.
func (n *NodesMetrics) ListNodesMetrics() map[string]*mv1beta1.NodeMetrics {
	return n.mx
}

// ListClusterMetrics collects total available cpu and mem on the cluster.
func (n *NodesMetrics) ListClusterMetrics(nmx map[string]*mv1beta1.NodeMetrics) v1.ResourceList {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	for _, mx := range nmx {
		cpu.Add(*mx.Usage.Cpu())
		mem.Add(*mx.Usage.Memory())
	}

	return v1.ResourceList{
		v1.ResourceCPU:    *cpu,
		v1.ResourceMemory: *mem,
	}
}
