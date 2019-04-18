package linter

import (
	"context"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment represents a Deployment linter.
type Deployment struct {
	*Linter
}

// NewDeployment returns a new Deployment linter.
func NewDeployment(l Loader, log *zerolog.Logger) *Deployment {
	return &Deployment{NewLinter(l, log)}
}

// Lint a Deployment.
func (d *Deployment) Lint(ctx context.Context) error {
	dps, err := d.ListDeployments()
	if err != nil {
		return err
	}

	d.lint(dps)

	return nil
}

func (d *Deployment) lint(dps map[string]appsv1.Deployment) {
	for fqn, dp := range dps {
		d.initIssues(fqn)
		d.checkDeployment(fqn, dp)
		d.checkContainers(fqn, dp)

		pmx := make(k8s.PodsMetrics)
		if err := podsMetrics(d, pmx); err != nil {
			continue
		}
		d.checkUtilization(fqn, dp, pmx)
	}
}

func (d *Deployment) checkDeployment(fqn string, dp appsv1.Deployment) {
	if dp.Spec.Replicas == nil || (dp.Spec.Replicas != nil && *dp.Spec.Replicas == 0) {
		d.addIssue(fqn, InfoLevel, "Zero scale detected")
	}

	if dp.Status.AvailableReplicas == 0 {
		d.addIssue(fqn, WarnLevel, "Used?")
	}

	if dp.Status.CollisionCount != nil {
		d.addIssuef(fqn, ErrorLevel, "ReplicaSet collisions detected %d", *dp.Status.CollisionCount)
	}
}

func (d *Deployment) checkContainers(fqn string, dp appsv1.Deployment) {
	spec := dp.Spec.Template.Spec

	l := NewContainer(d, d.log)
	for _, co := range spec.InitContainers {
		l.lint(co, false)
	}

	for _, co := range spec.Containers {
		l.lint(co, false)
	}

	d.addContainerIssues(fqn, l.Issues())
}

func (d *Deployment) checkUtilization(fqn string, dp appsv1.Deployment, pmx k8s.PodsMetrics) error {
	mx, err := deploymentUsage(d, dp, pmx)
	if err != nil {
		return err
	}

	// No resources bail!
	if mx.RequestedCPU.IsZero() && mx.RequestedMEM.IsZero() {
		return nil
	}

	cpuPerc := mx.ReqCPURatio()
	if cpuPerc > int64(d.CPUResourceLimits().Over) {
		d.addIssuef(fqn, WarnLevel, "CPU over allocated. Requested:%s - Current:%s (%s)", asMC(mx.RequestedCPU), asMC(mx.CurrentCPU), asPerc(cpuPerc))
	}

	if cpuPerc > 0 && cpuPerc < int64(d.CPUResourceLimits().Under) {
		d.addIssuef(fqn, WarnLevel, "CPU under allocated. Requested:%s - Current:%s (%s)", asMC(mx.RequestedCPU), asMC(mx.CurrentCPU), asPerc(cpuPerc))
	}

	memPerc := mx.ReqMEMRatio()
	if memPerc > int64(d.MEMResourceLimits().Over) {
		d.addIssuef(fqn, WarnLevel, "Memory over allocated. Requested:%s - Current:%s (%s)", asMB(mx.RequestedMEM), asMB(mx.CurrentMEM), asPerc(memPerc))
	}

	if memPerc > 0 && memPerc < int64(d.MEMResourceLimits().Under) {
		d.addIssuef(fqn, WarnLevel, "Memory under allocated. Requested:%s - Current:%s (%s)", asMB(mx.RequestedMEM), asMB(mx.CurrentMEM), asPerc(memPerc))
	}

	return nil
}

func deploymentUsage(l Loader, dp appsv1.Deployment, pmx k8s.PodsMetrics) (ConsumptionMetrics, error) {
	var mx ConsumptionMetrics

	sel, err := metav1.LabelSelectorAsSelector(dp.Spec.Selector)
	if err != nil {
		return mx, err
	}
	pods, err := l.ListPodsByLabels(sel.String())
	if err != nil {
		return mx, err
	}

	rc, rm := podResources(dp.Spec.Template.Spec)
	if dp.Spec.Replicas != nil {
		for i := 0; i < int(*dp.Spec.Replicas); i++ {
			mx.RequestedCPU.Add(rc)
			mx.RequestedMEM.Add(rm)
		}
	}

	for pfqn := range pods {
		if ccx, ok := pmx[pfqn]; ok {
			for _, cx := range ccx {
				mx.CurrentCPU.Add(cx.CurrentCPU)
				mx.CurrentMEM.Add(cx.CurrentMEM)
			}
		}
	}

	return mx, nil
}
