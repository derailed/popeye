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

func TestCRLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRole](ctx, l.DB, "auth/cr/1.yaml", internal.Glossary[internal.CR]))
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRoleBinding](ctx, l.DB, "auth/crb/1.yaml", internal.Glossary[internal.CRB]))
	assert.NoError(t, test.LoadDB[*rbacv1.RoleBinding](ctx, l.DB, "auth/rob/1.yaml", internal.Glossary[internal.ROB]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))

	cr := NewClusterRole(test.MakeCollector(t), dba)
	assert.Nil(t, cr.Lint(test.MakeContext("rbac.authorization.k8s.io/v1/clusterroles", "clusterroles")))
	assert.Equal(t, 3, len(cr.Outcome()))

	ii := cr.Outcome()["cr1"]
	assert.Equal(t, 0, len(ii))

	ii = cr.Outcome()["cr2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = cr.Outcome()["cr3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)
}

func TestCRLintAggregations(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRole](ctx, l.DB, "auth/cr/2.yaml", internal.Glossary[internal.CR]))

	cr := NewClusterRole(test.MakeCollector(t), dba)
	assert.Nil(t, cr.Lint(test.MakeContext("rbac.authorization.k8s.io/v1/clusterroles", "clusterroles")))
	assert.Equal(t, 2, len(cr.Outcome()))

	ii := cr.Outcome()["cr4"]
	assert.Equal(t, 1, len(ii))

	ii = cr.Outcome()["cr5"]
	assert.Equal(t, 0, len(ii))
}
