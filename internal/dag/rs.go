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

// ListReplicaSets list all included ReplicaSets.
func ListReplicaSets(ctx context.Context) (map[string]*appsv1.ReplicaSet, error) {
	rss, err := listAllReplicaSets(ctx)
	if err != nil {
		return map[string]*appsv1.ReplicaSet{}, err
	}

	f := mustExtractFactory(ctx)
	res := make(map[string]*appsv1.ReplicaSet, len(rss))
	for fqn, rs := range rss {
		if includeNS(f.Client(), rs.Namespace) {
			res[fqn] = rs
		}
	}

	return res, nil
}

// ListAllReplicaSets fetch all ReplicaSets on the cluster.
func listAllReplicaSets(ctx context.Context) (map[string]*appsv1.ReplicaSet, error) {
	ll, err := fetchReplicaSets(ctx)
	if err != nil {
		return nil, err
	}

	rss := make(map[string]*appsv1.ReplicaSet, len(ll.Items))
	for i := range ll.Items {
		rss[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return rss, nil
}

// FetchReplicaSets retrieves all ReplicaSets on the cluster.
func fetchReplicaSets(ctx context.Context) (*appsv1.ReplicaSetList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	dial, err := f.Client().Dial()
	if err != nil {
		return nil, err
	}
	if cfg.Flags.StandAlone {
		return dial.AppsV1().ReplicaSets(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("apps/v1/replicasets"))
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll appsv1.ReplicaSetList
	for _, o := range oo {
		var rs appsv1.ReplicaSet
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rs)
		if err != nil {
			return nil, errors.New("expecting replicaset resource")
		}
		ll.Items = append(ll.Items, rs)
	}

	return &ll, nil
}
