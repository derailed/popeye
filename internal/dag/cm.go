package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListConfigMaps list all included ConfigMaps.
func ListConfigMaps(c *k8s.Client, cfg *config.Config) (map[string]*v1.ConfigMap, error) {
	cms, err := listAllConfigMaps(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.ConfigMap, len(cms))
	for fqn, cm := range cms {
		if includeNS(c, cm.Namespace) {
			res[fqn] = cm
		}
	}

	return res, nil
}

// ListAllConfigMaps fetch all ConfigMaps on the cluster.
func listAllConfigMaps(c *k8s.Client) (map[string]*v1.ConfigMap, error) {
	ll, err := fetchConfigMaps(c)
	if err != nil {
		return nil, err
	}

	cms := make(map[string]*v1.ConfigMap, len(ll.Items))
	for i := range ll.Items {
		cms[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return cms, nil
}

// FetchConfigMaps retrieves all ConfigMaps on the cluster.
func fetchConfigMaps(c *k8s.Client) (*v1.ConfigMapList, error) {
	return c.DialOrDie().CoreV1().ConfigMaps(c.ActiveNamespace()).List(metav1.ListOptions{})
}
