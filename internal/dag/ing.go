package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	nv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListIngresses list all included Ingresses.
func ListIngresses(ctx context.Context) (map[string]*nv1beta1.Ingress, error) {
	ings, err := listAllIngresses(ctx)
	if err != nil {
		return map[string]*nv1beta1.Ingress{}, err
	}

	f := mustExtractFactory(ctx)
	res := make(map[string]*nv1beta1.Ingress, len(ings))
	for fqn, ing := range ings {
		if includeNS(f.Client(), ing.Namespace) {
			res[fqn] = ing
		}
	}

	return res, nil
}

// ListAllIngresses fetch all Ingresses on the cluster.
func listAllIngresses(ctx context.Context) (map[string]*nv1beta1.Ingress, error) {
	ll, err := fetchIngresses(ctx)
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
func fetchIngresses(ctx context.Context) (*nv1beta1.IngressList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	dial, err := f.Client().Dial()
	if err != nil {
		return nil, err
	}
	if cfg.Flags.StandAlone {
		return dial.ExtensionsV1beta1().Ingresses(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("extensions/v1beta1/ingresses"))
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
