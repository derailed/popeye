package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListStatefulSets list available StatefulSets.
func ListStatefulSets(ctx context.Context) (map[string]*appsv1.StatefulSet, error) {
	return listAllStatefulSets(ctx)
}

// ListAllStatefulSets fetch all StatefulSets on the cluster.
func listAllStatefulSets(ctx context.Context) (map[string]*appsv1.StatefulSet, error) {
	ll, err := fetchStatefulSets(ctx)
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
func fetchStatefulSets(ctx context.Context) (*appsv1.StatefulSetList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.AppsV1().StatefulSets(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("apps/v1/statefulsets"))
	oo, err := res.List(ctx)
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
