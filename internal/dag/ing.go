package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	nv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListIngresses list all included Ingresses.
func ListIngresses(f types.Factory, cfg *config.Config) (map[string]*nv1beta1.Ingress, error) {
	ings, err := listAllIngresses(f)
	if err != nil {
		return map[string]*nv1beta1.Ingress{}, err
	}

	res := make(map[string]*nv1beta1.Ingress, len(ings))
	for fqn, ing := range ings {
		if includeNS(f.Client(), ing.Namespace) {
			res[fqn] = ing
		}
	}

	return res, nil
}

// ListAllIngresses fetch all Ingresses on the cluster.
func listAllIngresses(f types.Factory) (map[string]*nv1beta1.Ingress, error) {
	ll, err := fetchIngresses(f)
	if err != nil {
		return nil, err
	}

	ings := make(map[string]*nv1beta1.Ingress, len(ll.Items))
	for i := range ll.Items {
		ings[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return ings, nil
}

// FetchIngresses retrieves all Ingresses on the cluster.
func fetchIngresses(f types.Factory) (*nv1beta1.IngressList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("extensions/v1beta1/ingresses"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll nv1beta1.IngressList
	for _, o := range oo {
		var ing nv1beta1.Ingress
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ing)
		if err != nil {
			return nil, errors.New("expecting ingress resource")
		}
		ll.Items = append(ll.Items, ing)
	}

	return &ll, nil
}
