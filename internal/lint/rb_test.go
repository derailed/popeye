// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/internal/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestRBLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*rbacv1.RoleBinding](ctx, l.DB, "auth/rob/1.yaml", internal.Glossary[internal.ROB]))
	assert.NoError(t, test.LoadDB[*rbacv1.Role](ctx, l.DB, "auth/ro/1.yaml", internal.Glossary[internal.RO]))
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRole](ctx, l.DB, "auth/cr/1.yaml", internal.Glossary[internal.CR]))
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRoleBinding](ctx, l.DB, "auth/crb/1.yaml", internal.Glossary[internal.CRB]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))

	rb := NewRoleBinding(test.MakeCollector(t), dba)
	assert.Nil(t, rb.Lint(test.MakeContext("rbac.authorization.k8s.io/v1/rolebindings", "rolebindings")))
	assert.Equal(t, 3, len(rb.Outcome()))

	ii := rb.Outcome()["default/rb1"]
	assert.Equal(t, 0, len(ii))

	ii = rb.Outcome()["default/rb2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1300] References a Role (default/r-bozo) which does not exist`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)

	ii = rb.Outcome()["default/rb3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1300] References a ClusterRole (cr-bozo) which does not exist`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
}

func TestRB_boundDefaultSA(t *testing.T) {
	uu := map[string]struct {
		roPath, robPath string
		crPath, crbPath string
		e               bool
	}{
		"happy": {
			roPath:  "auth/ro/1.yaml",
			robPath: "auth/rob/1.yaml",
			crPath:  "auth/cr/1.yaml",
			crbPath: "auth/crb/1.yaml",
		},
		"role-bound": {
			roPath:  "auth/ro/1.yaml",
			robPath: "auth/rob/2.yaml",
			crPath:  "auth/cr/1.yaml",
			crbPath: "auth/crb/1.yaml",
		},
		"cluster-role-bound": {
			roPath:  "auth/ro/1.yaml",
			robPath: "auth/rob/1.yaml",
			crPath:  "auth/cr/1.yaml",
			crbPath: "auth/crb/2.yaml",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			dba, err := test.NewTestDB()
			assert.NoError(t, err)
			l := db.NewLoader(dba)

			ctx := test.MakeCtx(t)
			assert.NoError(t, test.LoadDB[*rbacv1.RoleBinding](ctx, l.DB, u.robPath, internal.Glossary[internal.ROB]))
			assert.NoError(t, test.LoadDB[*rbacv1.Role](ctx, l.DB, u.roPath, internal.Glossary[internal.RO]))
			assert.NoError(t, test.LoadDB[*rbacv1.ClusterRole](ctx, l.DB, u.crPath, internal.Glossary[internal.CR]))
			assert.NoError(t, test.LoadDB[*rbacv1.ClusterRoleBinding](ctx, l.DB, u.crbPath, internal.Glossary[internal.CRB]))
			assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))

			assert.Equal(t, u.e, boundDefaultSA(dba))
		})
	}
}
