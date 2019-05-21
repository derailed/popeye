package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment represents a Deployment sanitizer.
type Deployment struct {
	*issues.Collector
	*cache.Deployment
	*cache.PodsMetrics
	*cache.Pod
	*config.Config

	client *k8s.Client
}

// NewDeployment return a new Deployment sanitizer.
func NewDeployment(c *k8s.Client, cfg *config.Config) Sanitizer {
	d := Deployment{client: c, Collector: issues.NewCollector(), Config: cfg}

	dps, err := dag.ListDeployments(c, cfg)
	if err != nil {
		d.AddErr("deployments", err)
	}
	d.Deployment = cache.NewDeployment(dps)

	mx, err := dag.ListPodsMetrics(c)
	if err != nil {
		d.AddInfof("podmetrics", "No metric-server detected %v", err)
	}
	d.PodsMetrics = cache.NewPodsMetrics(mx)

	pods, err := dag.ListPods(c, cfg)
	if err != nil {
		d.AddErr("pods", err)
	}
	d.Pod = cache.NewPod(pods)

	return &d
}

// Sanitize all available Deployments.
func (d *Deployment) Sanitize(ctx context.Context) error {
	return sanitize.NewDeployment(d.Collector, d).Sanitize(ctx)
}

// ListPodsByLabels retrieves all Pods matching a label selector in the allowed namespaces.
func (d *Deployment) ListPodsByLabels(sel string) (map[string]*v1.Pod, error) {
	pods, err := d.client.DialOrDie().CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: sel,
	})
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Pod, len(pods.Items))
	for _, po := range pods.Items {
		if d.client.IsActiveNamespace(po.Namespace) && !d.ExcludedNS(po.Namespace) {
			res[cache.MetaFQN(po.ObjectMeta)] = &po
		}
	}

	return res, nil
}
