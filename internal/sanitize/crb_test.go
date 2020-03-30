package sanitize

import (
	"sync"
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCRBSanitize(t *testing.T) {
	uu := map[string]struct {
		lister CRBLister
		key    string
		issues []config.ID
	}{
		"exists": {
			key:    "crb1",
			lister: makeCRBLister(crbOpts{name: "crb1", refKind: "ClusterRole", refName: "cr1"}),
		},
		"not_exists": {
			key:    "crb1",
			lister: makeCRBLister(crbOpts{name: "crb1", refKind: "ClusterRole", refName: "blah"}),
			issues: []config.ID{1300},
		},
	}

	ctx := makeContext("crbs", "crb")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := NewClusterRoleBinding(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, c.Sanitize(ctx))
			validateIssues(t, u.key, c.Outcome(), u.issues)
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type crbOpts struct {
	name, refKind, refName string
}

type crb struct {
	name string
	opts crbOpts
}

var _ CRBLister = (*crb)(nil)

func makeCRBLister(opts crbOpts) *crb {
	return &crb{name: "crb1", opts: opts}
}

func (c *crb) ListClusterRoleBindings() map[string]*rbacv1.ClusterRoleBinding {
	return map[string]*rbacv1.ClusterRoleBinding{
		c.opts.name: makeCRB(c.opts.name, c.opts.refKind, c.opts.refName),
	}
}

func (c *crb) ListClusterRoles() map[string]*rbacv1.ClusterRole {
	return map[string]*rbacv1.ClusterRole{
		"cr1": makeCR("cr1"),
		"cr2": makeCR("cr2"),
	}
}

func (c *crb) ListRoles() map[string]*rbacv1.Role {
	return map[string]*rbacv1.Role{
		"default/ro1": makeRO("ro1"),
	}
}

func (c *crb) ClusterRoleRefs(*sync.Map)        {}
func (c *crb) ClusterRoleBindingRefs(*sync.Map) {}

func makeCR(n string) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}
}
