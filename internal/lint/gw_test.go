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

func TestGatewayLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*gwv1.GatewayClass](ctx, l.DB, "net/gwc/1.yaml", internal.Glossary[internal.GWC]))
	assert.NoError(t, test.LoadDB[*gwv1.Gateway](ctx, l.DB, "net/gw/1.yaml", internal.Glossary[internal.GW]))

	gw := NewGateway(test.MakeCollector(t), dba)
	assert.Nil(t, gw.Lint(test.MakeContext("gateway.networking.k8s.io/v1/gateways", "gateways")))
	assert.Equal(t, 2, len(gw.Outcome()))

	ii := gw.Outcome()["default/gw1"]
	assert.Equal(t, 0, len(ii))

	ii = gw.Outcome()["default/gw2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-407] Gateway references GatewayClass "gwc-bozo" which does not exist`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	// assert.Equal(t, `[POP-407] Gateway references GatewayClass "gwc-bozo" which does not exist`, ii[0].Message)
	// assert.Equal(t, rules.ErrorLevel, ii[0].Level)
}
