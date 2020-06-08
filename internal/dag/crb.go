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

// ListClusterRoleBindings list included ClusterRoleBindings.
func ListClusterRoleBindings(ctx context.Context) (map[string]*rbacv1.ClusterRoleBinding, error) {
	return listAllClusterRoleBindings(ctx)
}

// ListAllClusterRoleBindings fetch all ClusterRoleBindings on the cluster.
func listAllClusterRoleBindings(ctx context.Context) (map[string]*rbacv1.ClusterRoleBinding, error) {
	ll, err := fetchClusterRoleBindings(ctx)
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
func fetchClusterRoleBindings(ctx context.Context) (*rbacv1.ClusterRoleBindingList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("rbac.authorization.k8s.io/v1/clusterrolebindings"))
	oo, err := res.List(ctx)
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
