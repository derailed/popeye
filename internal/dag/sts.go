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

// ListStatefulSets list available StatefulSets.
func ListStatefulSets(f types.Factory, cfg *config.Config) (map[string]*appsv1.StatefulSet, error) {
	sts, err := listAllStatefulSets(f)
	if err != nil {
		return map[string]*appsv1.StatefulSet{}, err
	}

	res := make(map[string]*appsv1.StatefulSet, len(sts))
	for fqn, st := range sts {
		if includeNS(f.Client(), st.Namespace) {
			res[fqn] = st
		}
	}

	return res, nil
}

// ListAllStatefulSets fetch all StatefulSets on the cluster.
func listAllStatefulSets(f types.Factory) (map[string]*appsv1.StatefulSet, error) {
	ll, err := fetchStatefulSets(f)
	if err != nil {
		return nil, err
	}

	sts := make(map[string]*appsv1.StatefulSet, len(ll.Items))
	for i := range ll.Items {
		sts[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return sts, nil
}

// FetchStatefulSets retrieves all StatefulSets on the cluster.
func fetchStatefulSets(f types.Factory) (*appsv1.StatefulSetList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("apps/v1/statefulsets"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll appsv1.StatefulSetList
	for _, o := range oo {
		var sts appsv1.StatefulSet
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sts)
		if err != nil {
			return nil, errors.New("expecting sts resource")
		}
		ll.Items = append(ll.Items, sts)
	}

	return &ll, nil
}
