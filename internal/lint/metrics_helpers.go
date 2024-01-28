// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ConsumptionMetrics tracks managed pods resource utilization.
type ConsumptionMetrics struct {
	QOS                    v1.PodQOSClass
	CurrentCPU, CurrentMEM resource.Quantity
	RequestCPU, RequestMEM resource.Quantity
	LimitCPU, LimitMEM     resource.Quantity
	RequestedStorage       resource.Quantity
}

// ReqAbsCPURatio returns absolute cpu ratio.
func (d *ConsumptionMetrics) ReqAbsCPURatio() float64 {
	if d.CurrentCPU.Cmp(d.RequestCPU) == 1 {
		return toMCRatio(d.CurrentCPU, d.RequestCPU)
	}
	return toMCRatio(d.RequestCPU, d.CurrentCPU)
}

// ReqCPURatio returns current cpu usage over requested percentage.
func (d *ConsumptionMetrics) ReqCPURatio() float64 {
	return toMCRatio(d.CurrentCPU, d.RequestCPU)
}

// c =100m r=300m 100/300 1/3 -> over allocated if ratio < 1
// c= 300m r=100m 300/100 3   -> under allocated if ratio > 1

// ReqAbsMEMRatio returns absolute mem  ratio.
func (d *ConsumptionMetrics) ReqAbsMEMRatio() float64 {
	if d.CurrentMEM.Cmp(d.RequestMEM) == 1 {
		return toMEMRatio(d.CurrentMEM, d.RequestMEM)
	}
	return toMEMRatio(d.RequestMEM, d.CurrentMEM)
}

// ReqMEMRatio returns current memory usage over requested percentage.
func (d *ConsumptionMetrics) ReqMEMRatio() float64 {
	return toMEMRatio(d.CurrentMEM, d.RequestMEM)
}

// LimitCPURatio returns current cpu usage over requested percentage.
func (d *ConsumptionMetrics) LimitCPURatio() float64 {
	return toMCRatio(d.CurrentCPU, d.LimitCPU)
}

// LimitMEMRatio returns current memory usage over requested percentage.
func (d *ConsumptionMetrics) LimitMEMRatio() float64 {
	return toMEMRatio(d.CurrentMEM, d.LimitMEM)
}
