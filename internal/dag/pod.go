package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPods list all filtered pods.
func ListPods(c *k8s.Client, cfg *config.Config) (map[string]*v1.Pod, error) {
	pods, err := listAllPods(c)
	if err != nil {
		return map[string]*v1.Pod{}, err
	}
	res := make(map[string]*v1.Pod, len(pods))
	for fqn, po := range pods {
		if includeNS(c, cfg, po.Namespace) && !cfg.ShouldExclude("pod", fqn) {
			res[fqn] = po
		}
	}

	return res, nil
}

// ListAllPods fetch all Pods on the cluster.
func listAllPods(c *k8s.Client) (map[string]*v1.Pod, error) {
	ll, err := fetchPods(c)
	if err != nil {
		log.Debug().Err(err).Msg("ListAll")
		return nil, err
	}

	pods := make(map[string]*v1.Pod, len(ll.Items))
	for i := range ll.Items {
		pods[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pods, nil
}

// FetchConfigMaps retrieves all ConfigMaps on the cluster.
func fetchPods(c *k8s.Client) (*v1.PodList, error) {
	return c.DialOrDie().CoreV1().Pods(c.ActiveNamespace()).List(metav1.ListOptions{})
}
