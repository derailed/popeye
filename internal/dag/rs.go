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

// ListReplicaSets list all included ReplicaSets.
func ListReplicaSets(f types.Factory, cfg *config.Config) (map[string]*appsv1.ReplicaSet, error) {
	rss, err := listAllReplicaSets(f)
	if err != nil {
		return map[string]*appsv1.ReplicaSet{}, err
	}

	res := make(map[string]*appsv1.ReplicaSet, len(rss))
	for fqn, rs := range rss {
		if includeNS(f.Client(), rs.Namespace) {
			res[fqn] = rs
		}
	}

	return res, nil
}

// ListAllReplicaSets fetch all ReplicaSets on the cluster.
func listAllReplicaSets(f types.Factory) (map[string]*appsv1.ReplicaSet, error) {
	ll, err := fetchReplicaSets(f)
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
func fetchReplicaSets(f types.Factory) (*appsv1.ReplicaSetList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("apps/v1/replicasets"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
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
