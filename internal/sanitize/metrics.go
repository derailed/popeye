package sanitize

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// ConsumptionMetrics tracks managed pods resource utilization.
type ConsumptionMetrics struct {
	CurrentCPU       resource.Quantity
	CurrentMEM       resource.Quantity
	RequestedCPU     resource.Quantity
	RequestedMEM     resource.Quantity
	RequestedStorage resource.Quantity
}

// ReqCPURatio returns current cpu usage over requested percentage.
func (d *ConsumptionMetrics) ReqCPURatio() int64 {
	return toMCRatio(d.RequestedCPU, d.CurrentCPU)
}

// ReqMEMRatio returns current memory usage over requested percentage.
func (d *ConsumptionMetrics) ReqMEMRatio() int64 {
	return toMEMRatio(d.RequestedMEM, d.CurrentMEM)
}
