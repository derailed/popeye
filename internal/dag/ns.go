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

// ListNamespaces list all included Namespaces.
func ListNamespaces(ctx context.Context) (map[string]*v1.Namespace, error) {
	return listAllNamespaces(ctx)
}

// ListAllNamespaces fetch all Namespaces on the cluster.
func listAllNamespaces(ctx context.Context) (map[string]*v1.Namespace, error) {
	ll, err := fetchNamespaces(ctx)
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
func fetchNamespaces(ctx context.Context) (*v1.NamespaceList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("v1/namespaces"))
	oo, err := res.List(ctx)
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
