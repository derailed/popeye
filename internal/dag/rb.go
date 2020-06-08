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

// ListRoleBindings list included RoleBindings.
func ListRoleBindings(ctx context.Context) (map[string]*rbacv1.RoleBinding, error) {
	return listAllRoleBindings(ctx)
}

// ListAllRoleBindings fetch all RoleBindings on the cluster.
func listAllRoleBindings(ctx context.Context) (map[string]*rbacv1.RoleBinding, error) {
	ll, err := fetchRoleBindings(ctx)
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
func fetchRoleBindings(ctx context.Context) (*rbacv1.RoleBindingList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.RbacV1().RoleBindings(f.Client().ActiveNamespace()).List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("rbac.authorization.k8s.io/v1/rolebindings"))
	oo, err := res.List(ctx)
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
