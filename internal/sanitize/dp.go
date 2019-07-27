package sanitize

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const utilFmt = "At current load, %s. Current:%s vs Requested:%s (%s)"

type (
	// PopeyeKey tracks context keys.
	PopeyeKey string

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
		AddSubCode(id issues.ID, p, s string, args ...interface{})
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
func (d *Deployment) Sanitize(ctx context.Context) error {
	over := pullOverAllocs(ctx)
	for fqn, dp := range d.ListDeployments() {
		d.InitOutcome(fqn)
		d.checkDeprecation(fqn, dp)
		d.checkDeployment(fqn, dp)
		d.checkContainers(fqn, dp.Spec.Template.Spec)
		pmx := k8s.PodsMetrics{}
		podsMetrics(d, pmx)

		d.checkUtilization(over, fqn, dp, pmx)
	}

	return nil
}

func (d *Deployment) checkDeprecation(fqn string, dp *appsv1.Deployment) {
	const current = "apps/v1"

	rev, err := resourceRev(fqn, dp.Annotations)
	if err != nil {
		rev = revFromLink(dp.SelfLink)
		if rev == "" {
			d.AddCode(404, fqn, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		d.AddCode(403, fqn, "Deployment", rev, current)
	}
}

// CheckDeployment checks if deployment contract is currently happy or not.
func (d *Deployment) checkDeployment(fqn string, dp *appsv1.Deployment) {
	if dp.Spec.Replicas == nil || (dp.Spec.Replicas != nil && *dp.Spec.Replicas == 0) {
		d.AddCode(500, fqn)
	}

	if dp.Status.AvailableReplicas == 0 {
		d.AddCode(501, fqn)
	}

	if dp.Status.CollisionCount != nil && *dp.Status.CollisionCount > 0 {
		d.AddCode(502, fqn, *dp.Status.CollisionCount)
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
func (d *Deployment) checkUtilization(over bool, fqn string, dp *appsv1.Deployment, pmx k8s.PodsMetrics) error {
	mx := d.deploymentUsage(dp, pmx)
	if mx.RequestCPU.IsZero() && mx.RequestMEM.IsZero() {
		return nil
	}

	cpuPerc := mx.ReqCPURatio()
	if cpuPerc > 1 && cpuPerc > float64(d.CPUResourceLimits().UnderPerc) {
		d.AddCode(503, fqn, asMC(mx.CurrentCPU), asMC(mx.RequestCPU), asPerc(cpuPerc))
	} else if over && cpuPerc < float64(d.CPUResourceLimits().OverPerc) {
		d.AddCode(504, fqn, asMC(mx.CurrentCPU), asMC(mx.RequestCPU), asPerc(mx.ReqAbsCPURatio()))
	}

	memPerc := mx.ReqMEMRatio()
	if memPerc > 1 && memPerc > float64(d.MEMResourceLimits().UnderPerc) {
		d.AddCode(505, fqn, asMB(mx.CurrentMEM), asMB(mx.RequestMEM), asPerc(memPerc))
	} else if over && memPerc < float64(d.MEMResourceLimits().OverPerc) {
		d.AddCode(506, fqn, asMB(mx.CurrentMEM), asMB(mx.RequestMEM), asPerc(mx.ReqAbsMEMRatio()))
	}

	return nil
}

// DeploymentUsage finds deployment running pods and compute current vs requested resource usage.
func (d *Deployment) deploymentUsage(dp *appsv1.Deployment, pmx k8s.PodsMetrics) ConsumptionMetrics {
	var mx ConsumptionMetrics
	for pfqn, pod := range d.ListPodsBySelector(dp.Spec.Selector) {
		cpu, mem := computePodResources(pod.Spec)
		mx.QOS = pod.Status.QOSClass
		mx.RequestCPU.Add(cpu)
		mx.RequestMEM.Add(mem)

		ccx, ok := pmx[pfqn]
		if !ok {
			continue
		}
		for _, cx := range ccx {
			mx.CurrentCPU.Add(cx.CurrentCPU)
			mx.CurrentMEM.Add(cx.CurrentMEM)
		}
	}

	return mx
}

// Helpers...

// PullOverAllocs check for over allocation setting in context.
func pullOverAllocs(ctx context.Context) bool {
	over := ctx.Value(PopeyeKey("OverAllocs"))
	if over == nil {
		return false
	}
	return over.(bool)
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

func computePodResources(spec v1.PodSpec) (cpu, mem resource.Quantity) {
	for _, co := range spec.InitContainers {
		c, m, _ := containerResources(co)
		if c != nil {
			cpu.Add(*c)
		}
		if m != nil {
			mem.Add(*m)
		}
	}

	for _, co := range spec.Containers {
		c, m, _ := containerResources(co)
		if c != nil {
			cpu.Add(*c)
		}
		if m != nil {
			mem.Add(*m)
		}
	}

	return
}

// ResourceRev is resource was deployed via kubectl check annotation for manifest rev.
func resourceRev(fqn string, a map[string]string) (string, error) {
	raw, ok := a["kubectl.kubernetes.io/last-applied-configuration"]
	if !ok {
		return "", fmt.Errorf("Raw resource manifest not available for %s", fqn)
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return "", err
	}
	return m["apiVersion"].(string), nil
}

// RevFromLink. extract resource version from selflink.
func revFromLink(link string) string {
	tokens := strings.Split(link, "/")
	if len(tokens) < 4 {
		return ""
	}
	if isVersion(tokens[2]) {
		return tokens[2]
	}
	return path.Join(tokens[2], tokens[3])
}

func isVersion(s string) bool {
	vers := []string{"v1", "v1beta1", "v1beta2", "v2beta1", "v2beta2"}
	for _, v := range vers {
		if s == v {
			return true
		}
	}
	return false
}
