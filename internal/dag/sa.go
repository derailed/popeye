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

// ListServiceAccounts list included ServiceAccounts.
func ListServiceAccounts(f types.Factory, cfg *config.Config) (map[string]*v1.ServiceAccount, error) {
	sas, err := listAllServiceAccounts(f)
	if err != nil {
		return map[string]*v1.ServiceAccount{}, err
	}

	res := make(map[string]*v1.ServiceAccount, len(sas))
	for fqn, sa := range sas {
		if includeNS(f.Client(), sa.Namespace) {
			res[fqn] = sa
		}
	}

	return res, nil
}

// ListAllServiceAccounts fetch all ServiceAccounts on the cluster.
func listAllServiceAccounts(f types.Factory) (map[string]*v1.ServiceAccount, error) {
	ll, err := fetchServiceAccounts(f)
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
func fetchServiceAccounts(f types.Factory) (*v1.ServiceAccountList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("v1/serviceaccounts"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
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
