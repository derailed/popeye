package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListHorizontalPodAutoscalers list all included HorizontalPodAutoscalers.
func ListHorizontalPodAutoscalers(c *k8s.Client, cfg *config.Config) (map[string]*autoscalingv1.HorizontalPodAutoscaler, error) {
	hpas, err := listAllHorizontalPodAutoscalers(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*autoscalingv1.HorizontalPodAutoscaler, len(hpas))
	for fqn, hpa := range hpas {
		if c.IsActiveNamespace(hpa.Namespace) && !cfg.ExcludedNS(hpa.Namespace) {
			res[fqn] = hpa
		}
	}

	return res, nil
}

// ListAllHorizontalPodAutoscalers fetch all HorizontalPodAutoscalers on the cluster.
func listAllHorizontalPodAutoscalers(c *k8s.Client) (map[string]*autoscalingv1.HorizontalPodAutoscaler, error) {
	ll, err := fetchHorizontalPodAutoscalers(c)
	if err != nil {
		return nil, err
	}

	hpas := make(map[string]*autoscalingv1.HorizontalPodAutoscaler, len(ll.Items))
	for i := range ll.Items {
		hpas[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return hpas, nil
}

// FetchHorizontalPodAutoscalers retrieves all HorizontalPodAutoscalers on the cluster.
func fetchHorizontalPodAutoscalers(c *k8s.Client) (*autoscalingv1.HorizontalPodAutoscalerList, error) {
	return c.DialOrDie().AutoscalingV1().HorizontalPodAutoscalers("").List(metav1.ListOptions{})
}
