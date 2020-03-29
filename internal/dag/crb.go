package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListClusterRoleBindings list included ClusterRoleBindings.
func ListClusterRoleBindings(f types.Factory, cfg *config.Config) (map[string]*rbacv1.ClusterRoleBinding, error) {
	crbs, err := listAllClusterRoleBindings(f)
	if err != nil {
		return map[string]*rbacv1.ClusterRoleBinding{}, err
	}
	res := make(map[string]*rbacv1.ClusterRoleBinding, len(crbs))
	for fqn, crb := range crbs {
		if includeNS(f.Client(), crb.Namespace) {
			res[fqn] = crb
		}
	}

	return res, nil
}

// ListAllClusterRoleBindings fetch all ClusterRoleBindings on the cluster.
func listAllClusterRoleBindings(f types.Factory) (map[string]*rbacv1.ClusterRoleBinding, error) {
	ll, err := fetchClusterRoleBindings(f)
	if err != nil {
		return nil, err
	}

	crbs := make(map[string]*rbacv1.ClusterRoleBinding, len(ll.Items))
	for i := range ll.Items {
		crbs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return crbs, nil
}

// FetchClusterRoleBindings retrieves all ClusterRoleBindings on the cluster.
func fetchClusterRoleBindings(f types.Factory) (*rbacv1.ClusterRoleBindingList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("rbac.authorization.k8s.io/v1/clusterrolebindings"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll rbacv1.ClusterRoleBindingList
	for _, o := range oo {
		var crb rbacv1.ClusterRoleBinding
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crb)
		if err != nil {
			return nil, errors.New("expecting clusterrolebinding resource")
		}
		ll.Items = append(ll.Items, crb)
	}

	return &ll, nil
}
