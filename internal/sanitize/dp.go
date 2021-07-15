package sanitize

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// DeploymentLister list available Deployments on a cluster.
type DeploymentLister interface {
	ListDeployments() map[string]*appsv1.Deployment
}

// DPLister represents deployments and deps listers.
type DPLister interface {
	PodLimiter
	PodsMetricsLister
	PodSelectorLister
	ConfigLister
	DeploymentLister
	ListServiceAccounts() map[string]*v1.ServiceAccount
}

// Deployment tracks Deployment sanitization.
type Deployment struct {
	*issues.Collector
	DPLister
}

// NewDeployment returns a new sanitizer.
func NewDeployment(co *issues.Collector, lister DPLister) *Deployment {
	return &Deployment{
		Collector: co,
		DPLister:  lister,
	}
}

// Sanitize cleanse the resource.
func (d *Deployment) Sanitize(ctx context.Context) error {
	over := pullOverAllocs(ctx)
	for fqn, dp := range d.ListDeployments() {
		d.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		d.checkDeprecation(ctx, dp)
		d.checkDeployment(ctx, dp)
		d.checkContainers(ctx, dp.Spec.Template.Spec)
		pmx := client.PodsMetrics{}
		podsMetrics(d, pmx)
		d.checkUtilization(ctx, over, dp, pmx)

		if d.NoConcerns(fqn) && d.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			d.ClearOutcome(fqn)
		}
	}

	return nil
}

func (d *Deployment) checkDeprecation(ctx context.Context, dp *appsv1.Deployment) {
	const current = "apps/v1"

	fqn := internal.MustExtractFQN(ctx)
	rev, err := resourceRev(fqn, "Deployment", dp.Annotations)
	if err != nil {
		if rev = revFromLink(dp.SelfLink); rev == "" {
			return
		}
	}
	if rev != current {
		d.AddCode(ctx, 403, "Deployment", rev, current)
	}
}

// CheckDeployment checks if deployment contract is currently happy or not.
func (d *Deployment) checkDeployment(ctx context.Context, dp *appsv1.Deployment) {
	if dp.Spec.Replicas == nil || (dp.Spec.Replicas != nil && *dp.Spec.Replicas == 0) {
		d.AddCode(ctx, 500)
		return
	}

	if dp.Spec.Replicas != nil && *dp.Spec.Replicas != dp.Status.AvailableReplicas {
		d.AddCode(ctx, 501, *dp.Spec.Replicas, dp.Status.AvailableReplicas)
	}

	if dp.Spec.Template.Spec.ServiceAccountName == "" {
		return
	}

	if _, ok := d.ListServiceAccounts()[client.FQN(dp.Namespace, dp.Spec.Template.Spec.ServiceAccountName)]; !ok {
		d.AddCode(ctx, 507, dp.Spec.Template.Spec.ServiceAccountName)
	}
}

// CheckContainers runs thru deployment template and checks pod configuration.
func (d *Deployment) checkContainers(ctx context.Context, spec v1.PodSpec) {
	c := NewContainer(internal.MustExtractFQN(ctx), d)
	for _, co := range spec.InitContainers {
		c.sanitize(ctx, co, false)
	}
	for _, co := range spec.Containers {
		c.sanitize(ctx, co, false)
	}
}

// CheckUtilization checks deployments requested resources vs current utilization.
func (d *Deployment) checkUtilization(ctx context.Context, over bool, dp *appsv1.Deployment, pmx client.PodsMetrics) {
	mx := d.deploymentUsage(dp, pmx)
	if mx.RequestCPU.IsZero() && mx.RequestMEM.IsZero() {
		return
	}
	checkCPU(ctx, d, over, mx)
	checkMEM(ctx, d, over, mx)
}

// DeploymentUsage finds deployment running pods and compute current vs requested resource usage.
func (d *Deployment) deploymentUsage(dp *appsv1.Deployment, pmx client.PodsMetrics) ConsumptionMetrics {
	var mx ConsumptionMetrics
	for pfqn, pod := range d.ListPodsBySelector(dp.Namespace, dp.Spec.Selector) {
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
	over := ctx.Value(internal.KeyOverAllocs)
	if over == nil {
		return false
	}
	return over.(bool)
}

// PodsMetrics gathers pod's container metrics from metrics server.
func podsMetrics(l PodsMetricsLister, pmx client.PodsMetrics) {
	for fqn, mx := range l.ListPodsMetrics() {
		cmx := client.ContainerMetrics{}
		podToContainerMetrics(mx, cmx)
		pmx[fqn] = cmx
	}
}

// PodToContainerMetrics gather pod's container metrics from metrics server.
func podToContainerMetrics(pmx *mv1beta1.PodMetrics, cmx client.ContainerMetrics) {
	for _, co := range pmx.Containers {
		cmx[co.Name] = client.Metrics{
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
func resourceRev(fqn string, kind string, a map[string]string) (string, error) {
	raw, ok := a["kubectl.kubernetes.io/last-applied-configuration"]
	if !ok {
		return "", fmt.Errorf("Raw resource manifest not available for %s", fqn)
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return "", err
	}
	if m["kind"] == kind {
		return m["apiVersion"].(string), nil
	}

	return "", errors.New("no matching resource kind")
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
