// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package client

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

type (
	// Metrics represent an aggregation of all pod containers metrics.
	Metrics struct {
		CurrentCPU resource.Quantity
		CurrentMEM resource.Quantity
	}

	// NodeMetrics describes raw node metrics.
	NodeMetrics struct {
		CurrentCPU   resource.Quantity
		CurrentMEM   resource.Quantity
		AvailableCPU resource.Quantity
		AvailableMEM resource.Quantity
		TotalCPU     resource.Quantity
		TotalMEM     resource.Quantity
	}

	// NodesMetrics tracks usage metrics per nodes.
	NodesMetrics map[string]NodeMetrics

	// PodsMetrics tracks usage metrics per pods.
	PodsMetrics map[string]ContainerMetrics

	// ContainerMetrics tracks container metrics
	ContainerMetrics map[string]Metrics
)

// Empty checks if we have any metrics.
func (n NodeMetrics) Empty() bool {
	return n == NodeMetrics{}
}

// Empty checks if we have any metrics.
func (m Metrics) Empty() bool {
	return m == Metrics{}
}
