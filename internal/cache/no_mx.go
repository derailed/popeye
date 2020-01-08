package cache

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// NodesMetrics represents a Node metrics cache.
type NodesMetrics struct {
	nmx map[string]*mv1beta1.NodeMetrics
}

// NewNodesMetrics returns new Node metrics cache.
func NewNodesMetrics(mx map[string]*mv1beta1.NodeMetrics) *NodesMetrics {
	return &NodesMetrics{nmx: mx}
}

// ListNodesMetrics returns all available NodeMetrics on the cluster.
func (n *NodesMetrics) ListNodesMetrics() map[string]*mv1beta1.NodeMetrics {
	return n.nmx
}

// ListAllocatedMetrics collects total used cpu and mem on the cluster.
func (n *NodesMetrics) ListAllocatedMetrics() v1.ResourceList {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	for _, mx := range n.nmx {
		cpu.Add(*mx.Usage.Cpu())
		mem.Add(*mx.Usage.Memory())
	}

	return v1.ResourceList{
		v1.ResourceCPU:    *cpu,
		v1.ResourceMemory: *mem,
	}
}

// ListAvailableMetrics return the total cluster available cpu/mem.
func (n *NodesMetrics) ListAvailableMetrics(nn map[string]*v1.Node) v1.ResourceList {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	for _, n := range nn {
		cpu.Add(*n.Status.Allocatable.Cpu())
		mem.Add(*n.Status.Allocatable.Memory())
	}
	used := n.ListAllocatedMetrics()
	cpu.Sub(*used.Cpu())
	mem.Sub(*used.Memory())

	return v1.ResourceList{
		v1.ResourceCPU:    *cpu,
		v1.ResourceMemory: *mem,
	}
}
