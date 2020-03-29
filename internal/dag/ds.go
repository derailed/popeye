package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListDaemonSets list all included DaemonSets.
func ListDaemonSets(f types.Factory, cfg *config.Config) (map[string]*appsv1.DaemonSet, error) {
	dps, err := listAllDaemonSets(f)
	if err != nil {
		return map[string]*appsv1.DaemonSet{}, err
	}

	res := make(map[string]*appsv1.DaemonSet, len(dps))
	for fqn, dp := range dps {
		if includeNS(f.Client(), dp.Namespace) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllDaemonSets fetch all DaemonSets on the cluster.
func listAllDaemonSets(f types.Factory) (map[string]*appsv1.DaemonSet, error) {
	ll, err := fetchDaemonSets(f)
	if err != nil {
		return nil, err
	}

	dps := make(map[string]*appsv1.DaemonSet, len(ll.Items))
	for i := range ll.Items {
		dps[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return dps, nil
}

// FetchDaemonSets retrieves all DaemonSets on the cluster.
func fetchDaemonSets(f types.Factory) (*appsv1.DaemonSetList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("apps/v1/daemonsets"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll appsv1.DaemonSetList
	for _, o := range oo {
		var ds appsv1.DaemonSet
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
		if err != nil {
			return nil, errors.New("expecting daemonset resource")
		}
		ll.Items = append(ll.Items, ds)
	}

	return &ll, nil
}
