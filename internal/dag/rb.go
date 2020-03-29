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

// ListRoleBindings list included RoleBindings.
func ListRoleBindings(f types.Factory, cfg *config.Config) (map[string]*rbacv1.RoleBinding, error) {
	rbs, err := listAllRoleBindings(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*rbacv1.RoleBinding, len(rbs))
	for fqn, rb := range rbs {
		if includeNS(f.Client(), rb.Namespace) {
			res[fqn] = rb
		}
	}

	return res, nil
}

// ListAllRoleBindings fetch all RoleBindings on the cluster.
func listAllRoleBindings(f types.Factory) (map[string]*rbacv1.RoleBinding, error) {
	ll, err := fetchRoleBindings(f)
	if err != nil {
		return nil, err
	}

	rbs := make(map[string]*rbacv1.RoleBinding, len(ll.Items))
	for i := range ll.Items {
		rbs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return rbs, nil
}

// FetchRoleBindings retrieves all RoleBindings on the cluster.
func fetchRoleBindings(f types.Factory) (*rbacv1.RoleBindingList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("rbac.authorization.k8s.io/v1/rolebindings"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll rbacv1.RoleBindingList
	for _, o := range oo {
		var rb rbacv1.RoleBinding
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rb)
		if err != nil {
			return nil, errors.New("expecting rolebinding resource")
		}
		ll.Items = append(ll.Items, rb)
	}

	return &ll, nil
}
