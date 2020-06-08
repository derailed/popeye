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

// ListServiceAccounts list included ServiceAccounts.
func ListServiceAccounts(ctx context.Context) (map[string]*v1.ServiceAccount, error) {
	return listAllServiceAccounts(ctx)
}

// ListAllServiceAccounts fetch all ServiceAccounts on the cluster.
func listAllServiceAccounts(ctx context.Context) (map[string]*v1.ServiceAccount, error) {
	ll, err := fetchServiceAccounts(ctx)
	if err != nil {
		return nil, err
	}
	sas := make(map[string]*v1.ServiceAccount, len(ll.Items))
	for i := range ll.Items {
		sas[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return sas, nil
}

// FetchServiceAccounts retrieves all ServiceAccounts on the cluster.
func fetchServiceAccounts(ctx context.Context) (*v1.ServiceAccountList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.CoreV1().ServiceAccounts(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("v1/serviceaccounts"))
	oo, err := res.List(ctx)
	if err != nil {
		return nil, err
	}
	var ll v1.ServiceAccountList
	for _, o := range oo {
		var sa v1.ServiceAccount
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sa)
		if err != nil {
			return nil, errors.New("expecting serviceaccount resource")
		}
		ll.Items = append(ll.Items, sa)
	}

	return &ll, nil

}
