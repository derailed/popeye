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

func TestCRBLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRoleBinding](ctx, l.DB, "auth/crb/1.yaml", internal.Glossary[internal.CRB]))
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRole](ctx, l.DB, "auth/cr/1.yaml", internal.Glossary[internal.CR]))
	assert.NoError(t, test.LoadDB[*rbacv1.Role](ctx, l.DB, "auth/ro/1.yaml", internal.Glossary[internal.RO]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))

	crb := NewClusterRoleBinding(test.MakeCollector(t), dba)
	assert.Nil(t, crb.Lint(test.MakeContext("rbac.authorization.k8s.io/v1/clusterrolebindings", "clusterrolebindings")))
	assert.Equal(t, 3, len(crb.Outcome()))

	ii := crb.Outcome()["crb1"]
	assert.Equal(t, 0, len(ii))

	ii = crb.Outcome()["crb2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1300] References a ClusterRole (cr-bozo) which does not exist`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)

	ii = crb.Outcome()["crb3"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-1300] References a Role (r-bozo) which does not exist`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-1300] References a ServiceAccount (default/sa-bozo) which does not exist`, ii[1].Message)
	assert.Equal(t, rules.WarnLevel, ii[1].Level)
}
