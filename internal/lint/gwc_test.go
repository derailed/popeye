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
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestGatewayClassLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*gwv1.GatewayClass](ctx, l.DB, "net/gwc/1.yaml", internal.Glossary[internal.GWC]))
	assert.NoError(t, test.LoadDB[*gwv1.Gateway](ctx, l.DB, "net/gw/1.yaml", internal.Glossary[internal.GW]))

	gwc := NewGatewayClass(test.MakeCollector(t), dba)
	assert.Nil(t, gwc.Lint(test.MakeContext("gateway.networking.k8s.io/v1/gatewayclasses", "gatewayclasses")))
	assert.Equal(t, 2, len(gwc.Outcome()))

	ii := gwc.Outcome()["gwc1"]
	assert.Equal(t, 0, len(ii))

	ii = gwc.Outcome()["gwc2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)
}
