package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// Deployment tracks Deployment sanitization.
	Deployment struct {
		*issues.Collector
		DeploymentLister
	}

	// PodsMetricsLister handles pods metrics.
	PodsMetricsLister interface {
		ListPodsMetrics() map[string]*mv1beta1.PodMetrics
	}

	// Collector collects sub issues.
	Collector interface {
		Outcome() issues.Outcome
		AddError(s, desc string)
		AddErrorf(s, fmat string, args ...interface{})
		AddSubOk(p, s, desc string)
		AddSubOkf(p, s, fmat string, args ...interface{})
		AddSubInfo(p, s, desc string)
		AddSubInfof(p, s, fmat string, args ...interface{})
		AddSubWarn(p, s, desc string)
		AddSubWarnf(p, s, fmat string, args ...interface{})
		AddSubError(p, s, desc string)
		AddSubErrorf(p, s, fmat string, args ...interface{})
	}

	// PodLimiter tracks metrics limit range.
	PodLimiter interface {
		PodCPULimit() float64
		PodMEMLimit() float64
		RestartsLimit() int
	}

	// PodSelectorLister list a collection of pod matching a selector.
	PodSelectorLister interface {
		ListPodsBySelector(sel *metav1.LabelSelector) map[string]*v1.Pod
	}

	// ConfigLister tracks configuration parameters.
	ConfigLister interface {
		CPUResourceLimits() config.Allocations
		MEMResourceLimits() config.Allocations
	}

	// DeployLister list deployments.
	DeployLister interface {
		ListDeployments() map[string]*appsv1.Deployment
	}

	// DeploymentLister list available Deployments on a cluster.
	DeploymentLister interface {
		PodLimiter
		PodsMetricsLister
		PodSelectorLister
		ConfigLister
		DeployLister
	}
)

// NewDeployment returns a new Deployment sanitizer.
func NewDeployment(co *issues.Collector, lister DeploymentLister) *Deployment {
	return &Deployment{
		Collector:        co,
		DeploymentLister: lister,
	}
}

// Sanitize configmaps.
func (d *Deployment) Sanitize(context.Context) error {
	for fqn, dp := range d.ListDeployments() {
		d.InitOutcome(fqn)

		d.checkDeployment(fqn, dp)
		d.checkContainers(fqn, dp.Spec.Template.Spec)
		pmx := k8s.PodsMetrics{}
		podsMetrics(d, pmx)
		d.checkUtilization(fqn, dp, pmx)
	}

	return nil
}

// CheckDeployment checks if deployment contract is currently happy or not.
func (d *Deployment) checkDeployment(fqn string, dp *appsv1.Deployment) {
	if dp.Spec.Replicas == nil || (dp.Spec.Replicas != nil && *dp.Spec.Replicas == 0) {
		d.AddInfo(fqn, "Zero scale detected")
	}

	if dp.Status.AvailableReplicas == 0 {
		d.AddWarn(fqn, "Used?")
	}

	if dp.Status.CollisionCount != nil && *dp.Status.CollisionCount > 0 {
		d.AddErrorf(fqn, "ReplicaSet collisions detected %d", *dp.Status.CollisionCount)
	}
}

// CheckContainers runs thru deployment template and checks pod configuration.
func (d *Deployment) checkContainers(fqn string, spec v1.PodSpec) {
	c := NewContainer(fqn, d)
	for _, co := range spec.InitContainers {
		c.sanitize(co, false)
	}
	for _, co := range spec.Containers {
		c.sanitize(co, false)
	}
}

// CheckUtilization checks deployments requested resources vs current utilization.
func (d *Deployment) checkUtilization(fqn string, dp *appsv1.Deployment, pmx k8s.PodsMetrics) error {
	mx, err := d.deploymentUsage(dp, pmx)
	if err != nil {
		return err
	}

	// No resources bail!
	if mx.RequestedCPU.IsZero() && mx.RequestedMEM.IsZero() {
		return nil
	}

	cpuPerc := mx.ReqCPURatio()
	if cpuPerc > int64(d.CPUResourceLimits().Over) {
		d.AddWarnf(fqn, "CPU over allocated. Current/Requested (%s/%s) ratio %s", asMC(mx.CurrentCPU), asMC(mx.RequestedCPU), asPerc(cpuPerc))
	}

	if cpuPerc > 0 && cpuPerc < int64(d.CPUResourceLimits().Under) {
		d.AddWarnf(fqn, "CPU under allocated. Current/Requested (%s/%s) ratio %s", asMC(mx.CurrentCPU), asMC(mx.RequestedCPU), asPerc(cpuPerc))
	}

	memPerc := mx.ReqMEMRatio()
	if memPerc > int64(d.MEMResourceLimits().Over) {
		d.AddWarnf(fqn, "Memory over allocated. Current/Requested (%s/%s) ratio %s", asMB(mx.CurrentMEM), asMB(mx.RequestedMEM), asPerc(memPerc))
	}

	if memPerc > 0 && memPerc < int64(d.MEMResourceLimits().Under) {
		d.AddWarnf(fqn, "Memory under allocated. Current/Requested (%s/%s) ratio %s", asMB(mx.CurrentMEM), asMB(mx.RequestedMEM), asPerc(memPerc))
	}

	return nil
}

// DeploymentUsage finds deployment running pods and compute current vs requested resource usage.
func (d *Deployment) deploymentUsage(dp *appsv1.Deployment, pmx k8s.PodsMetrics) (ConsumptionMetrics, error) {
	var mx ConsumptionMetrics
	rc, rm := podResources(dp.Spec.Template.Spec)
	if dp.Spec.Replicas != nil {
		for i := 0; i < int(*dp.Spec.Replicas); i++ {
			mx.RequestedCPU.Add(rc)
			mx.RequestedMEM.Add(rm)
		}
	}

	for pfqn := range d.ListPodsBySelector(dp.Spec.Selector) {
		if ccx, ok := pmx[pfqn]; ok {
			for _, cx := range ccx {
				mx.CurrentCPU.Add(cx.CurrentCPU)
				mx.CurrentMEM.Add(cx.CurrentMEM)
			}
		}
	}

	return mx, nil
}

// PodsMetrics gathers pod's container metrics from metrics server.
func podsMetrics(l PodsMetricsLister, pmx k8s.PodsMetrics) {
	for fqn, mx := range l.ListPodsMetrics() {
		cmx := k8s.ContainerMetrics{}
		podToContainerMetrics(mx, cmx)
		pmx[fqn] = cmx
	}
}

// PodToContainerMetrics gather pod's container metrics from metrics server.
func podToContainerMetrics(pmx *mv1beta1.PodMetrics, cmx k8s.ContainerMetrics) {
	for _, co := range pmx.Containers {
		cmx[co.Name] = k8s.Metrics{
			CurrentCPU: *co.Usage.Cpu(),
			CurrentMEM: *co.Usage.Memory(),
		}
	}
}
