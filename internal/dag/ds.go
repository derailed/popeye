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

// ListDaemonSets list all included DaemonSets.
func ListDaemonSets(ctx context.Context) (map[string]*appsv1.DaemonSet, error) {
	return listAllDaemonSets(ctx)
}

// ListAllDaemonSets fetch all DaemonSets on the cluster.
func listAllDaemonSets(ctx context.Context) (map[string]*appsv1.DaemonSet, error) {
	ll, err := fetchDaemonSets(ctx)
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
func fetchDaemonSets(ctx context.Context) (*appsv1.DaemonSetList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.AppsV1().DaemonSets(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("apps/v1/daemonsets"))
	oo, err := res.List(ctx)
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
