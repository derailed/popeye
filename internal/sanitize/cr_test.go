// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package sanitize

import (
	"regexp"
	"strconv"
	"sync"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCRSanitize(t *testing.T) {
	uu := map[string]struct {
		lister CRLister
		key    string
		issues []config.ID
	}{
		"usedCRBS": {
			key:    "cr1",
			lister: makeCRLister("cr1"),
		},
		"usedRBS": {
			key:    "cr2",
			lister: makeCRLister("cr2"),
		},
		"unused": {
			key:    "cr3",
			lister: makeCRLister("cr3"),
			issues: []config.ID{400},
		},
	}

	ctx := makeContext("rbac.authorization.k8s.io/v1/clusterroles", "cr")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := NewClusterRole(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, c.Sanitize(ctx))
			validateIssues(t, u.key, c.Outcome(), u.issues)
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

var issueRX = regexp.MustCompile(`\A\[POP-(\d+)\].`)

func validateIssues(t *testing.T, key string, actual issues.Outcome, expected []config.ID) {
	_, ok := actual[key]
	assert.True(t, ok, key)
	assert.Equal(t, len(expected), len(actual[key]))
	for _, id := range expected {
		a := actual[key]
		assert.Equal(t, 1, len(a))
		strs := issueRX.FindStringSubmatch(a[0].Message)
		assert.Equal(t, 2, len(strs))
		assert.Equal(t, strconv.Itoa(int(id)), strs[1])
		_ = id
	}
}

type crOpts struct {
	name, refKind, refName string
}

type cr struct {
	name string
	opts crOpts
}

var _ CRLister = (*cr)(nil)

func makeCRLister(n string) *cr {
	return &cr{name: n}
}

func (c *cr) ListClusterRoleBindings() map[string]*rbacv1.ClusterRoleBinding {
	return map[string]*rbacv1.ClusterRoleBinding{
		"default/crb1": makeCRB(c.opts.name, c.opts.refKind, c.opts.refName),
	}
}

func (c *cr) ListClusterRoles() map[string]*rbacv1.ClusterRole {
	return map[string]*rbacv1.ClusterRole{
		c.name: makeCR(c.name),
	}
}

func (c *cr) ListRoles() map[string]*rbacv1.Role {
	return map[string]*rbacv1.Role{
		"default/ro1": makeRO("ro1"),
	}
}

func (c *cr) ListRoleBindings() map[string]*rbacv1.RoleBinding {
	return map[string]*rbacv1.RoleBinding{
		"default/rb1": makeRB("rb1", "ClusterRole", "cr1"),
	}
}

func (c *cr) RoleRefs(refs *sync.Map) {
	refs.Store(cache.ResFqn(cache.ClusterRoleKey, "cr2"), internal.AllKeys)
}
func (c *cr) ClusterRoleRefs(refs *sync.Map) {
	refs.Store(cache.ResFqn(cache.ClusterRoleKey, "cr1"), internal.AllKeys)
}
func (c *cr) ClusterRoleBindingRefs(*sync.Map) {}

func makeRB(name, refKind, refName string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
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

func makeRO(n string) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}
}
