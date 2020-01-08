package sanitize

import (
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestROSanitize(t *testing.T) {
	uu := map[string]struct {
		lister ROLister
		key    string
		issues []config.ID
	}{
		"used": {
			key:    "default/ro1",
			lister: makeROLister("ro1", refOpts{refKind: "ClusterRole", refName: "cr1"}),
		},
		"unused": {
			key:    "default/ro3",
			lister: makeROLister("ro3", refOpts{refKind: "ClusterRole", refName: "cr1"}),
			issues: []config.ID{400},
		},
	}

	ctx := makeContext("roles")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			r := NewRole(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, r.Sanitize(ctx))
			validateIssues(t, u.key, r.Outcome(), u.issues)
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type refOpts struct {
	refKind, refName string
}

type ro struct {
	name string
	opts refOpts
}

var _ ROLister = (*ro)(nil)

func makeROLister(n string, opts refOpts) *ro {
	return &ro{name: n, opts: opts}
}

func (r *ro) ListRoleBindings() map[string]*rbacv1.RoleBinding {
	return map[string]*rbacv1.RoleBinding{
		"default/rb1": makeRB("rb1", "Role", r.name),
	}
}

func (r *ro) ListClusterRoleBindings() map[string]*rbacv1.ClusterRoleBinding {
	return map[string]*rbacv1.ClusterRoleBinding{
		"default/crb1": makeCRB("crb1", "ClusterRole", "cr2"),
	}
}

func (r *ro) ListClusterRoles() map[string]*rbacv1.ClusterRole {
	return map[string]*rbacv1.ClusterRole{
		"cr1": makeCR("cr1"),
	}
}

func (r *ro) ListRoles() map[string]*rbacv1.Role {
	return map[string]*rbacv1.Role{
		"default/" + r.name: makeRO(r.name),
	}
}

func (r *ro) ClusterRoleRefs(refs cache.ObjReferences) {
	refs[cache.ResFqn(cache.RoleKey, "default/ro2")] = internal.StringSet{"all": internal.Empty{}}
}
func (r *ro) RoleRefs(refs cache.ObjReferences) {
	refs[cache.ResFqn(cache.RoleKey, "default/ro1")] = internal.StringSet{"all": internal.Empty{}}
}

func makeCRB(name, refKind, refName string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: refKind,
			Name: refName,
		},
	}
}
