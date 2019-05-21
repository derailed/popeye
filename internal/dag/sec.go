package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListSecrets list all included Secrets.
func ListSecrets(c *k8s.Client, cfg *config.Config) (map[string]*v1.Secret, error) {
	secs, err := listAllSecrets(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Secret, len(secs))
	for fqn, sec := range secs {
		if includeNS(c, cfg, sec.Namespace) {
			res[fqn] = sec
		}
	}

	return res, nil
}

// ListAllSecrets fetch all Secrets on the cluster.
func listAllSecrets(c *k8s.Client) (map[string]*v1.Secret, error) {
	ll, err := fetchSecrets(c)
	if err != nil {
		return nil, err
	}

	secs := make(map[string]*v1.Secret, len(ll.Items))
	for i := range ll.Items {
		secs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return secs, nil
}

// FetchSecrets retrieves all Secrets on the cluster.
func fetchSecrets(c *k8s.Client) (*v1.SecretList, error) {
	return c.DialOrDie().CoreV1().Secrets("").List(metav1.ListOptions{})
}
