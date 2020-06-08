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

// ListClusterRoles list included ClusterRoles.
func ListClusterRoles(ctx context.Context) (map[string]*rbacv1.ClusterRole, error) {
	return listAllClusterRoles(ctx)
}

// ListAllClusterRoles fetch all ClusterRoles on the cluster.
func listAllClusterRoles(ctx context.Context) (map[string]*rbacv1.ClusterRole, error) {
	ll, err := fetchClusterRoles(ctx)
	if err != nil {
		return nil, err
	}
	crs := make(map[string]*rbacv1.ClusterRole, len(ll.Items))
	for i := range ll.Items {
		crs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return crs, nil
}

// FetchClusterRoles retrieves all ClusterRoles on the cluster.
func fetchClusterRoles(ctx context.Context) (*rbacv1.ClusterRoleList, error) {
	f, cfg := mustExtractFactory(ctx), mustExtractConfig(ctx)
	if cfg.Flags.StandAlone {
		dial, err := f.Client().Dial()
		if err != nil {
			return nil, err
		}
		return dial.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	}

	var res dao.Resource
	res.Init(f, client.NewGVR("rbac.authorization.k8s.io/v1/clusterroles"))
	oo, err := res.List(ctx)
	if err != nil {
		return nil, err
	}
	var ll rbacv1.ClusterRoleList
	for _, o := range oo {
		var cr rbacv1.ClusterRole
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cr)
		if err != nil {
			return nil, errors.New("expecting clusterrole resource")
		}
		ll.Items = append(ll.Items, cr)
	}

	return &ll, nil
}
