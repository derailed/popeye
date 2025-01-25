// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"testing"

	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cilium"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestCiliumNetworkPolicy(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v2.CiliumNetworkPolicy](ctx, l.DB, "cnp/1.yaml", internal.Glossary[cilium.CNP]))
	assert.NoError(t, test.LoadDB[*v2.CiliumEndpoint](ctx, l.DB, "cep/1.yaml", internal.Glossary[cilium.CEP]))

	li := NewCiliumNetworkPolicy(test.MakeCollector(t), dba)
	assert.Nil(t, li.Lint(test.MakeContext("cilium.io/v2/ciliumnetworkpolicies", "ciliumnetworkpolicies")))
	assert.Equal(t, 4, len(li.Outcome()))

	ii := li.Outcome()["default/cnp1"]
	assert.Equal(t, 0, len(ii))

	ii = li.Outcome()["default/cnp2"]
	assert.Equal(t, 3, len(ii))
	assert.Equal(t, `[POP-1700] No cilium endpoints matched endpoint selector`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-1700] No cilium endpoints matched ingress selector`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)
	assert.Equal(t, `[POP-1700] No cilium endpoints matched egress selector`, ii[2].Message)
	assert.Equal(t, rules.ErrorLevel, ii[2].Level)

	ii = li.Outcome()["default/cnp3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1700] No cilium endpoints matched endpoint selector`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)

	ii = li.Outcome()["default/cnp4"]
	assert.Equal(t, 0, len(ii))
}
