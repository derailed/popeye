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
)

func TestSecretLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.Secret](ctx, l.DB, "core/secret/1.yaml", internal.Glossary[internal.SEC]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*netv1.Ingress](ctx, l.DB, "net/ingress/1.yaml", internal.Glossary[internal.ING]))

	sec := NewSecret(test.MakeCollector(t), dba)
	assert.Nil(t, sec.Lint(test.MakeContext("v1/secrets", "secrets")))
	assert.Equal(t, 3, len(sec.Outcome()))

	ii := sec.Outcome()["default/sec1"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-401] Key "ns" used? Unable to locate key reference`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = sec.Outcome()["default/sec2"]
	assert.Equal(t, 0, len(ii))

	ii = sec.Outcome()["default/sec3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

}
