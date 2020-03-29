package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	nv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListNetworkPolicies list all included NetworkPolicies.
func ListNetworkPolicies(f types.Factory, cfg *config.Config) (map[string]*nv1.NetworkPolicy, error) {
	dps, err := listAllNetworkPolicies(f)
	if err != nil {
		return map[string]*nv1.NetworkPolicy{}, err
	}

	res := make(map[string]*nv1.NetworkPolicy, len(dps))
	for fqn, dp := range dps {
		if includeNS(f.Client(), dp.Namespace) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllNetworkPolicies fetch all NetworkPolicies on the cluster.
func listAllNetworkPolicies(f types.Factory) (map[string]*nv1.NetworkPolicy, error) {
	ll, err := fetchNetworkPolicies(f)
	if err != nil {
		return nil, err
	}

	dps := make(map[string]*nv1.NetworkPolicy, len(ll.Items))
	for i := range ll.Items {
		dps[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return dps, nil
}

// FetchNetworkPolicies retrieves all NetworkPolicies on the cluster.
func fetchNetworkPolicies(f types.Factory) (*nv1.NetworkPolicyList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("networking.k8s.io/v1/networkpolicies"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll nv1.NetworkPolicyList
	for _, o := range oo {
		var np nv1.NetworkPolicy
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &np)
		if err != nil {
			return nil, errors.New("expecting networkpolicy resource")
		}
		ll.Items = append(ll.Items, np)
	}

	return &ll, nil
}
