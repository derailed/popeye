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

// ListRoles list included Roles.
func ListRoles(f types.Factory, cfg *config.Config) (map[string]*rbacv1.Role, error) {
	ros, err := listAllRoles(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*rbacv1.Role, len(ros))
	for fqn, ro := range ros {
		if includeNS(f.Client(), ro.Namespace) {
			res[fqn] = ro
		}
	}

	return res, nil
}

// ListAllRoles fetch all Roles on the cluster.
func listAllRoles(f types.Factory) (map[string]*rbacv1.Role, error) {
	ll, err := fetchRoles(f)
	if err != nil {
		return nil, err
	}

	ros := make(map[string]*rbacv1.Role, len(ll.Items))
	for i := range ll.Items {
		ros[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return ros, nil
}

// FetchRoleBindings retrieves all RoleBindings on the cluster.
func fetchRoles(f types.Factory) (*rbacv1.RoleList, error) {
	var res dao.Resource
	res.Init(f, client.NewGVR("rbac.authorization.k8s.io/v1/roles"))

	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	oo, err := res.List(ctx, client.AllNamespaces)
	if err != nil {
		return nil, err
	}
	var ll rbacv1.RoleList
	for _, o := range oo {
		var ro rbacv1.Role
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ro)
		if err != nil {
			return nil, errors.New("expecting role resource")
		}
		ll.Items = append(ll.Items, ro)
	}

	return &ll, nil
}
