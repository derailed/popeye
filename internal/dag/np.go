package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	nv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListNetworkPolicies list all included NetworkPolicies.
func ListNetworkPolicies(ctx context.Context) (map[string]*nv1.NetworkPolicy, error) {
	return listAllNetworkPolicies(ctx)
}

// ListAllNetworkPolicies fetch all NetworkPolicies on the cluster.
func listAllNetworkPolicies(ctx context.Context) (map[string]*nv1.NetworkPolicy, error) {
	ll, err := fetchNetworkPolicies(ctx)
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
func fetchNetworkPolicies(ctx context.Context) (*nv1.NetworkPolicyList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.NetworkingV1().NetworkPolicies(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("networking.k8s.io/v1/networkpolicies"))
	oo, err := res.List(ctx)
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
