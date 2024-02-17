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
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestROLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*rbacv1.Role](ctx, l.DB, "auth/ro/1.yaml", internal.Glossary[internal.RO]))
	assert.NoError(t, test.LoadDB[*rbacv1.RoleBinding](ctx, l.DB, "auth/rob/1.yaml", internal.Glossary[internal.ROB]))
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRoleBinding](ctx, l.DB, "auth/crb/1.yaml", internal.Glossary[internal.CRB]))

	ro := NewRole(test.MakeCollector(t), dba)
	assert.Nil(t, ro.Lint(test.MakeContext("rbac.authorization.k8s.io/v1/roles", "roles")))
	assert.Equal(t, 3, len(ro.Outcome()))

	ii := ro.Outcome()["default/r1"]
	assert.Equal(t, 0, len(ii))

	ii = ro.Outcome()["default/r2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = ro.Outcome()["default/r3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)
}
