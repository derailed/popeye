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
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestHttpRouteTestLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*gwv1.HTTPRoute](ctx, l.DB, "net/gwr/1.yaml", internal.Glossary[internal.GWR]))
	assert.NoError(t, test.LoadDB[*gwv1.GatewayClass](ctx, l.DB, "net/gwc/1.yaml", internal.Glossary[internal.GWC]))
	assert.NoError(t, test.LoadDB[*gwv1.Gateway](ctx, l.DB, "net/gw/1.yaml", internal.Glossary[internal.GW]))
	assert.NoError(t, test.LoadDB[*v1.Service](ctx, l.DB, "core/svc/1.yaml", internal.Glossary[internal.SVC]))

	hr := NewHTTPRoute(test.MakeCollector(t), dba)
	assert.Nil(t, hr.Lint(test.MakeContext("gateway.networking.k8s.io/v1/httproutes", "httproutes")))
	assert.Equal(t, 3, len(hr.Outcome()))

	ii := hr.Outcome()["default/r1"]
	assert.Equal(t, 0, len(ii))

	ii = hr.Outcome()["default/r2"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-407] HTTPRoute references Gateway "default/gw-bozo" which does not exist`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-1106] No target ports match service port 8080`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)

	ii = hr.Outcome()["default/r3"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-407] HTTPRoute references Service "default/svc-bozo" which does not exist`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-1106] No target ports match service port 9090`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)

}
