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

func TestIngLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*netv1.Ingress](ctx, l.DB, "net/ingress/1.yaml", internal.Glossary[internal.ING]))
	assert.NoError(t, test.LoadDB[*v1.Service](ctx, l.DB, "core/svc/1.yaml", internal.Glossary[internal.SVC]))

	ing := NewIngress(test.MakeCollector(t), dba)
	assert.Nil(t, ing.Lint(test.MakeContext("networking.k8s.io/v1/ingresses", "ingresses")))
	assert.Equal(t, 6, len(ing.Outcome()))

	ii := ing.Outcome()["default/ing1"]
	assert.Equal(t, 0, len(ii))

	ii = ing.Outcome()["default/ing2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1403] Ingress backend uses a port#, prefer a named port: 9090`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = ing.Outcome()["default/ing3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1401] Ingress references a service backend which does not exist: s2`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)

	ii = ing.Outcome()["default/ing4"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-1402] Ingress references a service port which is not defined: :0`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-1404] Invalid Ingress backend spec. Must use port name or number`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)

	ii = ing.Outcome()["default/ing5"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-1400] Ingress LoadBalancer port reported an error: boom`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-666] Lint internal error: Ingress local obj refs not supported`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)

	ii = ing.Outcome()["default/ing6"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-1402] Ingress references a service port which is not defined: :9091`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-1403] Ingress backend uses a port#, prefer a named port: 9091`, ii[1].Message)
	assert.Equal(t, rules.InfoLevel, ii[1].Level)
}
