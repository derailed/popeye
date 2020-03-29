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

// ListLimitRanges list all included LimitRanges.
func ListLimitRanges(f types.Factory, cfg *config.Config) (map[string]*v1.LimitRange, error) {
	lrs, err := listAllLimitRanges(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.LimitRange, len(lrs))
	for fqn, lr := range lrs {
		if includeNS(f.Client(), lr.Namespace) {
			res[fqn] = lr
		}
	}

	return res, nil
}

// ListAllLimitRanges fetch all LimitRanges on the cluster.
func listAllLimitRanges(f types.Factory) (map[string]*v1.LimitRange, error) {
	ll, err := fetchLimitRanges(f)
	if err != nil {
		return nil, err
	}

	lrs := make(map[string]*v1.LimitRange, len(ll.Items))
	for i := range ll.Items {
		lrs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return lrs, nil
}

// fetchLimitRanges retrieves all LimitRanges on the cluster.
func fetchLimitRanges(f types.Factory) (*v1.LimitRangeList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/limitranges"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll v1.LimitRangeList
	for _, o := range oo {
		var lr v1.LimitRange
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &lr)
		if err != nil {
			return nil, errors.New("expecting limitrange resource")
		}
		ll.Items = append(ll.Items, lr)
	}

	return &ll, nil
}
