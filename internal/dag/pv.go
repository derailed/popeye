package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListPersistentVolumes list all included PersistentVolumes.
func ListPersistentVolumes(f types.Factory, cfg *config.Config) (map[string]*v1.PersistentVolume, error) {
	pvs, err := listAllPersistentVolumes(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.PersistentVolume, len(pvs))
	for fqn, pv := range pvs {
		res[fqn] = pv
	}

	return res, nil
}

// ListAllPersistentVolumes fetch all PersistentVolumes on the cluster.
func listAllPersistentVolumes(f types.Factory) (map[string]*v1.PersistentVolume, error) {
	ll, err := fetchPersistentVolumes(f)
	if err != nil {
		return nil, err
	}

	pvs := make(map[string]*v1.PersistentVolume, len(ll.Items))
	for i := range ll.Items {
		pvs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pvs, nil
}

// FetchPersistentVolumes retrieves all PersistentVolumes on the cluster.
func fetchPersistentVolumes(f types.Factory) (*v1.PersistentVolumeList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/persistentvolumes"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll v1.PersistentVolumeList
	for _, o := range oo {
		var pv v1.PersistentVolume
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pv)
		if err != nil {
			return nil, errors.New("expecting persistentvolume resource")
		}
		ll.Items = append(ll.Items, pv)
	}

	return &ll, nil
}
