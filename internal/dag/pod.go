package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPods list all filtered pods.
func ListPods(c *k8s.Client, cfg *config.Config) (map[string]*v1.Pod, error) {
	pods, err := listAllPods(c)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*v1.Pod, len(pods))
	for fqn, po := range pods {
		if c.IsActiveNamespace(po.Namespace) && !cfg.ExcludedNS(po.Namespace) {
			res[fqn] = po
		}
	}

	return res, nil
}

// ListAllPods fetch all Pods on the cluster.
func listAllPods(c *k8s.Client) (map[string]*v1.Pod, error) {
	ll, err := c.DialOrDie().CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods := make(map[string]*v1.Pod, len(ll.Items))
	for i := range ll.Items {
		pods[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pods, nil
}

// ListPodsByLabels retrieves all Pods matching a label selector in the allowed namespaces.
func ListPodsByLabels(c k8s.Client, cfg *config.Config, sel string) (map[string]*v1.Pod, error) {
	pods, err := c.DialOrDie().CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: sel,
	})
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Pod, len(pods.Items))
	for _, po := range pods.Items {
		if c.IsActiveNamespace(po.Namespace) && !cfg.ExcludedNS(po.Namespace) {
			res[metaFQN(po.ObjectMeta)] = &po
		}
	}

	return res, nil
}
