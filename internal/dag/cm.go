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

// ListConfigMaps list all included ConfigMaps.
func ListConfigMaps(f types.Factory, cfg *config.Config) (map[string]*v1.ConfigMap, error) {
	cms, err := listAllConfigMaps(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.ConfigMap, len(cms))
	for fqn, cm := range cms {
		if includeNS(f.Client(), cm.Namespace) {
			res[fqn] = cm
		}
	}

	return res, nil
}

// ListAllConfigMaps fetch all ConfigMaps on the cluster.
func listAllConfigMaps(f types.Factory) (map[string]*v1.ConfigMap, error) {
	ll, err := fetchConfigMaps(f)
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
func fetchConfigMaps(f types.Factory) (*v1.ConfigMapList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/configmaps"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll v1.ConfigMapList
	for _, o := range oo {
		var cm v1.ConfigMap
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cm)
		if err != nil {
			return nil, errors.New("expecting configmap resource")
		}
		ll.Items = append(ll.Items, cm)
	}

	return &ll, nil
}
