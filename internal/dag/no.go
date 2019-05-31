package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNodes list all included Nodes.
func ListNodes(c *k8s.Client, cfg *config.Config) (map[string]*v1.Node, error) {
	nos, err := listAllNodes(c)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*v1.Node, len(nos))
	for fqn, no := range nos {
		if !cfg.ShouldExclude("node", fqn) {
			res[fqn] = no
		}
	}

	return res, nil
}

// ListAllNodes fetch all Nodes on the cluster.
func listAllNodes(c *k8s.Client) (map[string]*v1.Node, error) {
	ll, err := fetchNodes(c)
	if err != nil {
		return nil, err
	}

	nos := make(map[string]*v1.Node, len(ll.Items))
	for i := range ll.Items {
		nos[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return nos, nil
}

// FetchNodes retrieves all Nodes on the cluster.
func fetchNodes(c *k8s.Client) (*v1.NodeList, error) {
	return c.DialOrDie().CoreV1().Nodes().List(metav1.ListOptions{})
}
