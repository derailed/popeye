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

// ListPersistentVolumeClaims list all included PersistentVolumeClaims.
func ListPersistentVolumeClaims(f types.Factory, cfg *config.Config) (map[string]*v1.PersistentVolumeClaim, error) {
	pvcs, err := listAllPersistentVolumeClaims(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.PersistentVolumeClaim, len(pvcs))
	for fqn, pvc := range pvcs {
		if includeNS(f.Client(), pvc.Namespace) {
			res[fqn] = pvc
		}
	}

	return res, nil
}

// ListAllPersistentVolumeClaims fetch all PersistentVolumeClaims on the cluster.
func listAllPersistentVolumeClaims(f types.Factory) (map[string]*v1.PersistentVolumeClaim, error) {
	ll, err := fetchPersistentVolumeClaims(f)
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
func fetchPersistentVolumeClaims(f types.Factory) (*v1.PersistentVolumeClaimList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/persistentvolumeclaims"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
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
