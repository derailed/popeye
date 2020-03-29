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

// ListEndpoints list all included Endpoints.
func ListEndpoints(f types.Factory, cfg *config.Config) (map[string]*v1.Endpoints, error) {
	eps, err := listAllEndpoints(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Endpoints, len(eps))
	for fqn, ep := range eps {
		if includeNS(f.Client(), ep.Namespace) {
			res[fqn] = ep
		}
	}

	return res, nil
}

// ListAllEndpoints fetch all Endpoints on the cluster.
func listAllEndpoints(f types.Factory) (map[string]*v1.Endpoints, error) {
	ll, err := fetchEndpoints(f)
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
func fetchEndpoints(f types.Factory) (*v1.EndpointsList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/endpoints"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
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
