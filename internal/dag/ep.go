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

// ListEndpoints list all included Endpoints.
func ListEndpoints(ctx context.Context) (map[string]*v1.Endpoints, error) {
	return listAllEndpoints(ctx)
}

// ListAllEndpoints fetch all Endpoints on the cluster.
func listAllEndpoints(ctx context.Context) (map[string]*v1.Endpoints, error) {
	ll, err := fetchEndpoints(ctx)
	if err != nil {
		return nil, err
	}
	eps := make(map[string]*v1.Endpoints, len(ll.Items))
	for i := range ll.Items {
		eps[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return eps, nil
}

// FetchEndpoints retrieves all Endpoints on the cluster.
func fetchEndpoints(ctx context.Context) (*v1.EndpointsList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.CoreV1().Endpoints(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("v1/endpoints"))
	oo, err := res.List(ctx)
	if err != nil {
		return nil, err
	}
	var ll v1.EndpointsList
	for _, o := range oo {
		var ep v1.Endpoints
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ep)
		if err != nil {
			return nil, errors.New("expecting endpoints resource")
		}
		ll.Items = append(ll.Items, ep)
	}

	return &ll, nil

}
