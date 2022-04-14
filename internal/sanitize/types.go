package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// Collector collects sub issues.
type Collector interface {
	// Outcome collects issues.
	Outcome() issues.Outcome

	// AddSubCode records a sub issue.
	AddSubCode(ctx context.Context, id config.ID, args ...interface{})

	// AddCode records a new issue.
	AddCode(ctx context.Context, id config.ID, args ...interface{})
}

// PodsMetricsLister handles pods metrics.
type PodsMetricsLister interface {
	ListPodsMetrics() map[string]*mv1beta1.PodMetrics
}

// PodLimiter tracks metrics limit range.
type PodLimiter interface {
	PodCPULimit() float64
	PodMEMLimit() float64
	RestartsLimit() int
}

type ContainerRestrictor interface {
	AllowedRegistries() []string
}

// PodSelectorLister list a collection of pod matching a selector.
type PodSelectorLister interface {
	ListPodsBySelector(ns string, sel *metav1.LabelSelector) map[string]*v1.Pod
}

// ConfigLister tracks configuration parameters.
type ConfigLister interface {
	// CPUResourceLimits returns the CPU utilization threshold.
	CPUResourceLimits() config.Allocations

	// MEMResourceLimits returns the MEM utilization threshold.
	MEMResourceLimits() config.Allocations
}
