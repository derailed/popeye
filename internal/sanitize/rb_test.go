package sanitize

import (
	"sync"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestRBSanitize(t *testing.T) {
	uu := map[string]struct {
		lister RBLister
		key    string
		issues []config.ID
	}{
		"used": {
			key:    "default/rb1",
			lister: makeRBLister(rbOpts{name: "rb1", refKind: "Role", refName: "r1"}),
		},
		"unused": {
			key:    "default/rb1",
			lister: makeRBLister(rbOpts{name: "rb1", refKind: "Role", refName: "blah"}),
			issues: []config.ID{1300},
		},
	}

	ctx := makeContext("rbac.authorization.k8s.io/v1/rolebindings", "rb")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			r := NewRoleBinding(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, r.Sanitize(ctx))
			validateIssues(t, u.key, r.Outcome(), u.issues)
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type rbOpts struct {
	name, refKind, refName string
}

type rb struct {
	opts rbOpts
}

var _ RBLister = (*rb)(nil)

func makeRBLister(opts rbOpts) *rb {
	return &rb{opts: opts}
}

func (r *rb) ListRoleBindings() map[string]*rbacv1.RoleBinding {
	return map[string]*rbacv1.RoleBinding{
		"default/" + r.opts.name: makeRB(r.opts.name, r.opts.refKind, r.opts.refName),
	}
}

func (r *rb) ListClusterRoles() map[string]*rbacv1.ClusterRole {
	return map[string]*rbacv1.ClusterRole{
		"cr1": makeCR("cr1"),
	}
}

func (r *rb) ListRoles() map[string]*rbacv1.Role {
	return map[string]*rbacv1.Role{
		"default/r1": makeRO("r1"),
	}
}

func (r *rb) RoleRefs(refs *sync.Map) {
	refs.Store("default/ro1", internal.AllKeys)
}
