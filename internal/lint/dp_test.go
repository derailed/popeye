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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestDPLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*appsv1.Deployment](ctx, l.DB, "apps/dp/1.yaml", internal.Glossary[internal.DP]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*mv1beta1.PodMetrics](ctx, l.DB, "mx/pod/1.yaml", internal.Glossary[internal.PMX]))

	dp := NewDeployment(test.MakeCollector(t), dba)
	assert.Nil(t, dp.Lint(test.MakeContext("apps/v1/deployments", "deployments")))
	assert.Equal(t, 3, len(dp.Outcome()))

	ii := dp.Outcome()["default/dp1"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-503] At current load, CPU under allocated. Current:20000m vs Requested:1000m (2000.00%)`, ii[0].Message)
	assert.Equal(t, `[POP-505] At current load, Memory under allocated. Current:20Mi vs Requested:1Mi (2000.00%)`, ii[1].Message)

	ii = dp.Outcome()["default/dp2"]
	assert.Equal(t, 5, len(ii))
	assert.Equal(t, `[POP-501] Unhealthy 1 desired but have 0 available`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-507] Deployment references ServiceAccount "sa-bozo" which does not exist`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)
	assert.Equal(t, `[POP-106] No resources requests/limits defined`, ii[2].Message)
	assert.Equal(t, rules.WarnLevel, ii[2].Level)
	assert.Equal(t, `[POP-108] Unnamed port 3000`, ii[3].Message)
	assert.Equal(t, rules.InfoLevel, ii[3].Level)
	assert.Equal(t, `[POP-508] No pods match controller selector: app=pod-bozo`, ii[4].Message)
	assert.Equal(t, rules.ErrorLevel, ii[4].Level)

	ii = dp.Outcome()["default/dp3"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-500] Zero scale detected`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-666] Lint internal error: no pod selector given`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)
}
