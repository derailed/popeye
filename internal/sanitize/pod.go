package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	v1 "k8s.io/api/core/v1"
	pv1beta1 "k8s.io/api/policy/v1beta1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	// SecNonRootUndefined denotes no root user set
	SecNonRootUndefined NonRootUser = iota - 1
	// SecNonRootUnset denotes root user
	SecNonRootUnset = 0
	// SecNonRootSet denotes non root user
	SecNonRootSet = 1
)

// NonRootUser identifies if a security context for nonRootUser is set/unset or undefined.
type NonRootUser int

type (
	// Pod represents a Pod linter.
	Pod struct {
		*issues.Collector
		PodMXLister
	}

	// PdbLister list pdb matching a given selector
	PdbLister interface {
		ListPodDisruptionBudgets() map[string]*pv1beta1.PodDisruptionBudget
		ForLabels(labels map[string]string) *pv1beta1.PodDisruptionBudget
	}

	// PodLister lists available pods.
	PodLister interface {
		ListPods() map[string]*v1.Pod
		GetPod(sel map[string]string) *v1.Pod
	}

	// PodMXLister list available pods.
	PodMXLister interface {
		PodLimiter
		PodMetricsLister
		PodLister
		PdbLister
		ConfigLister
	}

	// PodMetric tracks node metrics available and current range.
	PodMetric interface {
		CurrentCPU() int64
		CurrentMEM() int64
		Empty() bool
	}
)

// NewPod returns a new sanitizer.
func NewPod(co *issues.Collector, lister PodMXLister) *Pod {
	return &Pod{
		Collector:   co,
		PodMXLister: lister,
	}
}

// Sanitize cleanse the resource..
func (p *Pod) Sanitize(ctx context.Context) error {
	mx := p.ListPodsMetrics()
	for fqn, po := range p.ListPods() {
		p.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		p.checkStatus(ctx, po)
		p.checkContainerStatus(ctx, po)
		p.checkContainers(ctx, po)
		p.checkPdb(ctx, po.ObjectMeta.Labels)
		p.checkSecure(ctx, po.Spec)
		pmx, cmx := mx[fqn], k8s.ContainerMetrics{}
		containerMetrics(pmx, cmx)
		p.checkUtilization(ctx, po, cmx)

		if p.Config.ExcludeFQN(internal.MustExtractSection(ctx), fqn) {
			p.ClearOutcome(fqn)
		}
	}
	return nil
}

func (p *Pod) checkPdb(ctx context.Context, labels map[string]string) {
	if p.ForLabels(labels) == nil {
		p.AddCode(ctx, 206)
	}
}

func (p *Pod) checkUtilization(ctx context.Context, po *v1.Pod, cmx k8s.ContainerMetrics) {
	if len(cmx) == 0 {
		return
	}

	for _, co := range po.Spec.Containers {
		cmx, ok := cmx[co.Name]
		if !ok {
			continue
		}
		NewContainer(internal.MustExtractFQN(ctx), p).checkUtilization(ctx, co, cmx)
	}
}

func (p *Pod) checkSecure(ctx context.Context, spec v1.PodSpec) {
	if spec.ServiceAccountName == "default" {
		p.AddCode(ctx, 300)
	}

	if spec.AutomountServiceAccountToken == nil || *spec.AutomountServiceAccountToken {
		p.AddCode(ctx, 301)
	}

	if spec.SecurityContext == nil {
		return
	}

	// If pod security ctx is present and we have
	podSec := hasPodNonRootUser(spec.SecurityContext)
	var victims int
	for _, co := range spec.InitContainers {
		if !checkCOSecurityContext(co) && !podSec {
			victims++
			p.AddSubCode(internal.WithGroup(ctx, co.Name), 306)
		}
	}
	for _, co := range spec.Containers {
		if !checkCOSecurityContext(co) && !podSec {
			victims++
			p.AddSubCode(internal.WithGroup(ctx, co.Name), 306)
		}
	}
	if victims > 0 && !podSec {
		p.AddCode(ctx, 302)
	}
}

func checkCOSecurityContext(co v1.Container) bool {
	return hasCoNonRootUser(co.SecurityContext)
}

func hasPodNonRootUser(sec *v1.PodSecurityContext) bool {
	if sec == nil {
		return false
	}
	if sec.RunAsNonRoot != nil {
		return *sec.RunAsNonRoot
	}
	if sec.RunAsUser != nil {
		return *sec.RunAsUser != 0
	}
	return false
}

func hasCoNonRootUser(sec *v1.SecurityContext) bool {
	if sec == nil {
		return false
	}
	if sec.RunAsNonRoot != nil {
		return *sec.RunAsNonRoot
	}
	if sec.RunAsUser != nil {
		return *sec.RunAsUser != 0
	}
	return false
}

func (p *Pod) checkContainers(ctx context.Context, po *v1.Pod) {
	co := NewContainer(internal.MustExtractFQN(ctx), p)
	for _, c := range po.Spec.InitContainers {
		co.sanitize(ctx, c, false)
	}
	for _, c := range po.Spec.Containers {
		co.sanitize(ctx, c, !isPartOfJob(po))
	}
}

func (p *Pod) checkContainerStatus(ctx context.Context, po *v1.Pod) {
	limit := p.RestartsLimit()
	for _, s := range po.Status.InitContainerStatuses {
		cs := newContainerStatus(p, internal.MustExtractFQN(ctx), len(po.Status.InitContainerStatuses), true, limit)
		cs.sanitize(ctx, s)
	}

	for _, s := range po.Status.ContainerStatuses {
		cs := newContainerStatus(p, internal.MustExtractFQN(ctx), len(po.Status.ContainerStatuses), false, limit)
		cs.sanitize(ctx, s)
	}
}

func (p *Pod) checkStatus(ctx context.Context, po *v1.Pod) {
	switch po.Status.Phase {
	case v1.PodRunning:
	case v1.PodSucceeded:
	default:
		p.AddCode(ctx, 207, po.Status.Phase)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func containerMetrics(pmx *mv1beta1.PodMetrics, mx k8s.ContainerMetrics) {
	// No metrics -> Bail!
	if pmx == nil {
		return
	}

	for _, co := range pmx.Containers {
		mx[co.Name] = k8s.Metrics{
			CurrentCPU: *co.Usage.Cpu(),
			CurrentMEM: *co.Usage.Memory(),
		}
	}
}

func isPartOfJob(po *v1.Pod) bool {
	for _, o := range po.OwnerReferences {
		if o.Kind == "Job" {
			return true
		}
	}

	return false
}
