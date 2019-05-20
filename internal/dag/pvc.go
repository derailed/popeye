package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPersistentVolumeClaims list all included PersistentVolumeClaims.
func ListPersistentVolumeClaims(c *k8s.Client, cfg *config.Config) (map[string]*v1.PersistentVolumeClaim, error) {
	secs, err := listAllPersistentVolumeClaims(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.PersistentVolumeClaim, len(secs))
	for fqn, sec := range secs {
		if c.IsActiveNamespace(sec.Namespace) && !cfg.ExcludedNS(sec.Namespace) {
			res[fqn] = sec
		}
	}

	return res, nil
}

// ListAllPersistentVolumeClaims fetch all PersistentVolumeClaims on the cluster.
func listAllPersistentVolumeClaims(c *k8s.Client) (map[string]*v1.PersistentVolumeClaim, error) {
	ll, err := fetchPersistentVolumeClaims(c)
	if err != nil {
		return nil, err
	}

	secs := make(map[string]*v1.PersistentVolumeClaim, len(ll.Items))
	for i := range ll.Items {
		secs[MetaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return secs, nil
}

// FetchPersistentVolumeClaims retrieves all PersistentVolumeClaims on the cluster.
func fetchPersistentVolumeClaims(c *k8s.Client) (*v1.PersistentVolumeClaimList, error) {
	return c.DialOrDie().CoreV1().PersistentVolumeClaims("").List(metav1.ListOptions{})
}
