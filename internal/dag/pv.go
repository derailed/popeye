package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPersistentVolumes list all included PersistentVolumes.
func ListPersistentVolumes(c *k8s.Client, cfg *config.Config) (map[string]*v1.PersistentVolume, error) {
	pvs, err := listAllPersistentVolumes(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.PersistentVolume, len(pvs))
	for fqn, pv := range pvs {
		if !cfg.ShouldExclude("persistentvolume", fqn) {
			res[fqn] = pv
		}
	}

	return res, nil
}

// ListAllPersistentVolumes fetch all PersistentVolumes on the cluster.
func listAllPersistentVolumes(c *k8s.Client) (map[string]*v1.PersistentVolume, error) {
	ll, err := fetchPersistentVolumes(c)
	if err != nil {
		log.Debug().Err(err).Msg("ListAll")
		return nil, err
	}

	pvs := make(map[string]*v1.PersistentVolume, len(ll.Items))
	for i := range ll.Items {
		pvs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pvs, nil
}

// FetchPersistentVolumes retrieves all PersistentVolumes on the cluster.
func fetchPersistentVolumes(c *k8s.Client) (*v1.PersistentVolumeList, error) {
	return c.DialOrDie().CoreV1().PersistentVolumes().List(metav1.ListOptions{})
}
