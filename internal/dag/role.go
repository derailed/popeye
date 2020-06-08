package dag

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ListRoles list included Roles.
func ListRoles(ctx context.Context) (map[string]*rbacv1.Role, error) {
	return listAllRoles(ctx)
}

// ListAllRoles fetch all Roles on the cluster.
func listAllRoles(ctx context.Context) (map[string]*rbacv1.Role, error) {
	ll, err := fetchRoles(ctx)
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
func fetchRoles(ctx context.Context) (*rbacv1.RoleList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.RbacV1().Roles(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("rbac.authorization.k8s.io/v1/roles"))
	oo, err := res.List(ctx)
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
