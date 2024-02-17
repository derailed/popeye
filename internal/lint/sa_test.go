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
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestSALint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/2.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*rbacv1.RoleBinding](ctx, l.DB, "auth/rob/1.yaml", internal.Glossary[internal.ROB]))
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRoleBinding](ctx, l.DB, "auth/crb/1.yaml", internal.Glossary[internal.CRB]))
	assert.NoError(t, test.LoadDB[*v1.Secret](ctx, l.DB, "core/secret/1.yaml", internal.Glossary[internal.SEC]))
	assert.NoError(t, test.LoadDB[*v1.Service](ctx, l.DB, "core/svc/1.yaml", internal.Glossary[internal.SVC]))
	assert.NoError(t, test.LoadDB[*netv1.Ingress](ctx, l.DB, "net/ingress/1.yaml", internal.Glossary[internal.ING]))

	sa := NewServiceAccount(test.MakeCollector(t), dba)
	assert.Nil(t, sa.Lint(test.MakeContext("v1/serviceaccounts", "serviceaccounts")))
	assert.Equal(t, 6, len(sa.Outcome()))

	ii := sa.Outcome()["default/default"]
	assert.Equal(t, 0, len(ii))

	ii = sa.Outcome()["default/sa1"]
	assert.Equal(t, 0, len(ii))

	ii = sa.Outcome()["default/sa2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-303] Do you mean it? ServiceAccount is automounting APIServer credentials`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)

	ii = sa.Outcome()["default/sa3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-303] Do you mean it? ServiceAccount is automounting APIServer credentials`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)

	ii = sa.Outcome()["default/sa4"]
	assert.Equal(t, 3, len(ii))
	assert.Equal(t, `[POP-304] References a secret "default/bozo" which does not exist`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-305] References a pull secret which does not exist: default/s1`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[2].Message)
	assert.Equal(t, rules.InfoLevel, ii[2].Level)

	ii = sa.Outcome()["default/sa5"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-304] References a secret "default/s1" which does not exist`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-305] References a pull secret which does not exist: default/bozo`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)

}
