// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"github.com/derailed/popeye/internal/db"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ListAllocatedMetrics collects total used cpu and mem on the cluster.
func listAllocatedMetrics(db *db.DB) (v1.ResourceList, error) {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	mm, err := db.ListNMX()
	if err != nil {
		return nil, err
	}
	for _, mx := range mm {
		cpu.Add(*mx.Usage.Cpu())
		mem.Add(*mx.Usage.Memory())
	}

	return v1.ResourceList{v1.ResourceCPU: *cpu, v1.ResourceMemory: *mem}, nil
}

// ListAvailableMetrics return the total cluster available cpu/mem.
func ListAvailableMetrics(db *db.DB) (v1.ResourceList, error) {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	nn, err := db.ListNodes()
	if err != nil {
		return nil, err
	}
	for _, n := range nn {
		cpu.Add(*n.Status.Allocatable.Cpu())
		mem.Add(*n.Status.Allocatable.Memory())
	}
	used, err := listAllocatedMetrics(db)
	if err != nil {
		return nil, err
	}
	cpu.Sub(*used.Cpu())
	mem.Sub(*used.Memory())

	return v1.ResourceList{
		v1.ResourceCPU:    *cpu,
		v1.ResourceMemory: *mem,
	}, nil
}
