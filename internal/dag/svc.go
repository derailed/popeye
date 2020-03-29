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

// ListServices list all included Services.
func ListServices(f types.Factory, cfg *config.Config) (map[string]*v1.Service, error) {
	svcs, err := listAllServices(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Service, len(svcs))
	for fqn, svc := range svcs {
		if includeNS(f.Client(), svc.Namespace) {
			res[fqn] = svc
		}
	}

	return res, nil
}

// ListAllServices fetch all Services on the cluster.
func listAllServices(f types.Factory) (map[string]*v1.Service, error) {
	ll, err := fetchServices(f)
	if err != nil {
		return nil, err
	}

	svcs := make(map[string]*v1.Service, len(ll.Items))
	for i := range ll.Items {
		svcs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return svcs, nil
}

// FetchServices retrieves all Services on the cluster.
func fetchServices(f types.Factory) (*v1.ServiceList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/services"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll v1.ServiceList
	for _, o := range oo {
		var svc v1.Service
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &svc)
		if err != nil {
			return nil, errors.New("expecting service resource")
		}
		ll.Items = append(ll.Items, svc)
	}

	return &ll, nil
}
