package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListPersistentVolumes list all included PersistentVolumes.
func ListPersistentVolumes(ctx context.Context) (map[string]*v1.PersistentVolume, error) {
	return listAllPersistentVolumes(ctx)
}

// ListAllPersistentVolumes fetch all PersistentVolumes on the cluster.
func listAllPersistentVolumes(ctx context.Context) (map[string]*v1.PersistentVolume, error) {
	ll, err := fetchPersistentVolumes(ctx)
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
func fetchPersistentVolumes(ctx context.Context) (*v1.PersistentVolumeList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("v1/persistentvolumes"))
	oo, err := res.List(ctx)
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
