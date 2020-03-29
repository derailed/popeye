package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	pv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListPodSecurityPolicies list all included PodSecurityPolicies.
func ListPodSecurityPolicies(f types.Factory, cfg *config.Config) (map[string]*pv1beta1.PodSecurityPolicy, error) {
	dps, err := listAllPodSecurityPolicys(f)
	if err != nil {
		return map[string]*pv1beta1.PodSecurityPolicy{}, err
	}

	res := make(map[string]*pv1beta1.PodSecurityPolicy, len(dps))
	for fqn, dp := range dps {
		if includeNS(f.Client(), dp.Namespace) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllPodSecurityPolicys fetch all PodSecurityPolicys on the cluster.
func listAllPodSecurityPolicys(f types.Factory) (map[string]*pv1beta1.PodSecurityPolicy, error) {
	ll, err := fetchPodSecurityPolicys(f)
	if err != nil {
		return nil, err
	}

	dps := make(map[string]*pv1beta1.PodSecurityPolicy, len(ll.Items))
	for i := range ll.Items {
		dps[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return dps, nil
}

// FetchPodSecurityPolicys retrieves all PodSecurityPolicys on the cluster.
func fetchPodSecurityPolicys(f types.Factory) (*pv1beta1.PodSecurityPolicyList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("policy/v1beta1/podsecuritypolicies"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll pv1beta1.PodSecurityPolicyList
	for _, o := range oo {
		var psp pv1beta1.PodSecurityPolicy
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &psp)
		if err != nil {
			return nil, errors.New("expecting configmap resource")
		}
		ll.Items = append(ll.Items, psp)
	}

	return &ll, nil

}
