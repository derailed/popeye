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

// ListPersistentVolumeClaims list all included PersistentVolumeClaims.
func ListPersistentVolumeClaims(ctx context.Context) (map[string]*v1.PersistentVolumeClaim, error) {
	return listAllPersistentVolumeClaims(ctx)
}

// ListAllPersistentVolumeClaims fetch all PersistentVolumeClaims on the cluster.
func listAllPersistentVolumeClaims(ctx context.Context) (map[string]*v1.PersistentVolumeClaim, error) {
	ll, err := fetchPersistentVolumeClaims(ctx)
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
func fetchPersistentVolumeClaims(ctx context.Context) (*v1.PersistentVolumeClaimList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.CoreV1().PersistentVolumeClaims(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("v1/persistentvolumeclaims"))
	oo, err := res.List(ctx)
	if err != nil {
		return nil, err
	}
	var ll v1.PersistentVolumeClaimList
	for _, o := range oo {
		var pvc v1.PersistentVolumeClaim
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pvc)
		if err != nil {
			return nil, errors.New("expecting persistentvolumeclaim resource")
		}
		ll.Items = append(ll.Items, pvc)
	}

	return &ll, nil
}
