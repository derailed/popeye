package dag

import (
	"context"
	"errors"
	netv1 "k8s.io/api/networking/v1"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// IngressGVR tracks ingress specification
var IngressGVR = client.NewGVR("networking.k8s.io/v1/ingresses")

// ListIngresses list all included Ingresses.
func ListIngresses(ctx context.Context) (map[string]*netv1.Ingress, error) {
	return listAllIngresses(ctx)
}

// ListAllIngresses fetch all Ingresses on the cluster.
func listAllIngresses(ctx context.Context) (map[string]*netv1.Ingress, error) {
	ll, err := fetchIngresses(ctx)
	if err != nil {
		return nil, err
	}
	ings := make(map[string]*netv1.Ingress, len(ll.Items))
	for i := range ll.Items {
		ings[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return ings, nil
}

// FetchIngresses retrieves all Ingresses on the cluster.
func fetchIngresses(ctx context.Context) (*netv1.IngressList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}

		return dial.NetworkingV1().Ingresses(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, IngressGVR)
	oo, err := res.List(ctx)
	if err != nil {
		return nil, err
	}
	var ll netv1.IngressList
	for _, o := range oo {
		var ing netv1.Ingress
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ing)
		if err != nil {
			return nil, errors.New("expecting ingress resource")
		}
		ll.Items = append(ll.Items, ing)
	}

	return &ll, nil
}
