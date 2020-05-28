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

// ctx := context.WithValue(context.Background(), internal.KeyFactory, f)

// ListConfigMaps list all included ConfigMaps.
func ListConfigMaps(ctx context.Context) (map[string]*v1.ConfigMap, error) {
	cms, err := listAllConfigMaps(ctx)
	if err != nil {
		return nil, err
	}

	f := mustExtractFactory(ctx)
	res := make(map[string]*v1.ConfigMap, len(cms))
	for fqn, cm := range cms {
		if includeNS(f.Client(), cm.Namespace) {
			res[fqn] = cm
		}
	}

	return res, nil
}

// ListAllConfigMaps fetch all ConfigMaps on the cluster.
func listAllConfigMaps(ctx context.Context) (map[string]*v1.ConfigMap, error) {
	ll, err := fetchConfigMaps(ctx)
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
func fetchConfigMaps(ctx context.Context) (*v1.ConfigMapList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	dial, err := f.Client().Dial()
	if err != nil {
		return nil, err
	}
	if cfg.Flags.StandAlone {
		return dial.CoreV1().ConfigMaps(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("v1/configmaps"))
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
