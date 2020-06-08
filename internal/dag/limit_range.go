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

// ListLimitRanges list all included LimitRanges.
func ListLimitRanges(ctx context.Context) (map[string]*v1.LimitRange, error) {
	return listAllLimitRanges(ctx)
}

// ListAllLimitRanges fetch all LimitRanges on the cluster.
func listAllLimitRanges(ctx context.Context) (map[string]*v1.LimitRange, error) {
	ll, err := fetchLimitRanges(ctx)
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
func fetchLimitRanges(ctx context.Context) (*v1.LimitRangeList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.CoreV1().LimitRanges(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("v1/limitranges"))
	oo, err := res.List(ctx)
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
