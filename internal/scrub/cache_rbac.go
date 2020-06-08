package scrub

import (
	"context"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dag"
)

type rbac struct {
	*dial

	mx  sync.Mutex
	crb *cache.ClusterRoleBinding
	cr  *cache.ClusterRole
	rb  *cache.RoleBinding
	ro  *cache.Role
}

func newRBAC(d *dial) *rbac {
	return &rbac{dial: d}
}

func (r *rbac) roles() (*cache.Role, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.ro != nil {
		return r.ro, nil
	}
	ctx, cancel := r.context()
	defer cancel()
	ros, err := dag.ListRoles(ctx)
	r.ro = cache.NewRole(ros)

	return r.ro, err
}

func (r *rbac) rolebindings() (*cache.RoleBinding, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.rb != nil {
		return r.rb, nil
	}
	ctx, cancel := r.context()
	defer cancel()
	rbs, err := dag.ListRoleBindings(ctx)
	r.rb = cache.NewRoleBinding(rbs)

	return r.rb, err
}

func (r *rbac) clusterroles() (*cache.ClusterRole, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.cr != nil {
		return r.cr, nil
	}
	ctx, cancel := r.context()
	defer cancel()
	crs, err := dag.ListClusterRoles(ctx)
	r.cr = cache.NewClusterRole(crs)

	return r.cr, err
}

func (r *rbac) clusterrolebindings() (*cache.ClusterRoleBinding, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.crb != nil {
		return r.crb, nil
	}
	ctx, cancel := r.context()
	defer cancel()
	crbs, err := dag.ListClusterRoleBindings(ctx)
	r.crb = cache.NewClusterRoleBinding(crbs)

	return r.crb, err
}

// Helpers...

func (r *rbac) context() (context.Context, context.CancelFunc) {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, r.factory)
	ctx = context.WithValue(ctx, internal.KeyConfig, r.config)
	if r.config.Flags.ActiveNamespace != nil {
		ctx = context.WithValue(ctx, internal.KeyNamespace, *r.config.Flags.ActiveNamespace)
	} else {
		ns, err := r.factory.Client().Config().CurrentNamespaceName()
		if err != nil {
			ns = client.AllNamespaces
		}
		ctx = context.WithValue(ctx, internal.KeyNamespace, ns)
	}

	return context.WithCancel(ctx)
}
