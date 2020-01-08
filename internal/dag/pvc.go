package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPersistentVolumeClaims list all included PersistentVolumeClaims.
func ListPersistentVolumeClaims(c *k8s.Client, cfg *config.Config) (map[string]*v1.PersistentVolumeClaim, error) {
	pvcs, err := listAllPersistentVolumeClaims(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.PersistentVolumeClaim, len(pvcs))
	for fqn, pvc := range pvcs {
		if includeNS(c, pvc.Namespace) {
			res[fqn] = pvc
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

	pvcs := make(map[string]*v1.PersistentVolumeClaim, len(ll.Items))
	for i := range ll.Items {
		pvcs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pvcs, nil
}

// FetchPersistentVolumeClaims retrieves all PersistentVolumeClaims on the cluster.
func fetchPersistentVolumeClaims(c *k8s.Client) (*v1.PersistentVolumeClaimList, error) {
	return c.DialOrDie().CoreV1().PersistentVolumeClaims(c.ActiveNamespace()).List(metav1.ListOptions{})
}
