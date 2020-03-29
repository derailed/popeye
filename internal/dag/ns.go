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

// ListNamespaces list all included Namespaces.
func ListNamespaces(f types.Factory, cfg *config.Config) (map[string]*v1.Namespace, error) {
	nss, err := listAllNamespaces(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Namespace, len(nss))
	for fqn, ns := range nss {
		if includeNS(f.Client(), ns.Name) {
			res[fqn] = ns
		}
	}

	return res, nil
}

// ListAllNamespaces fetch all Namespaces on the cluster.
func listAllNamespaces(f types.Factory) (map[string]*v1.Namespace, error) {
	ll, err := fetchNamespaces(f)
	if err != nil {
		return nil, err
	}

	nss := make(map[string]*v1.Namespace, len(ll.Items))
	for i := range ll.Items {
		nss[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return nss, nil
}

// FetchNamespaces retrieves all Namespaces on the cluster.
func fetchNamespaces(f types.Factory) (*v1.NamespaceList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/namespaces"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll v1.NamespaceList
	for _, o := range oo {
		var ns v1.Namespace
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ns)
		if err != nil {
			return nil, errors.New("expecting namespace resource")
		}
		ll.Items = append(ll.Items, ns)
	}

	return &ll, nil
}
