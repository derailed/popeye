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

func TestDSLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*appsv1.DaemonSet](ctx, l.DB, "apps/ds/1.yaml", internal.Glossary[internal.DS]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*mv1beta1.PodMetrics](ctx, l.DB, "mx/pod/1.yaml", internal.Glossary[internal.PMX]))

	ds := NewDaemonSet(test.MakeCollector(t), dba)
	assert.Nil(t, ds.Lint(test.MakeContext("apps/v1/daemonsets", "daemonsets")))
	assert.Equal(t, 2, len(ds.Outcome()))

	ii := ds.Outcome()["default/ds1"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-503] At current load, CPU under allocated. Current:20000m vs Requested:1000m (2000.00%)`, ii[0].Message)
	assert.Equal(t, `[POP-505] At current load, Memory under allocated. Current:20Mi vs Requested:1Mi (2000.00%)`, ii[1].Message)

	ii = ds.Outcome()["default/ds2"]
	assert.Equal(t, 6, len(ii))
	assert.Equal(t, `[POP-507] Deployment references ServiceAccount "sa-bozo" which does not exist`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-100] Untagged docker image in use`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)
	assert.Equal(t, `[POP-106] No resources requests/limits defined`, ii[2].Message)
	assert.Equal(t, rules.WarnLevel, ii[2].Level)
	assert.Equal(t, `[POP-100] Untagged docker image in use`, ii[3].Message)
	assert.Equal(t, rules.ErrorLevel, ii[3].Level)
	assert.Equal(t, `[POP-106] No resources requests/limits defined`, ii[4].Message)
	assert.Equal(t, rules.WarnLevel, ii[4].Level)
	assert.Equal(t, `[POP-508] No pods match controller selector: app=p10`, ii[5].Message)
	assert.Equal(t, rules.ErrorLevel, ii[5].Level)
}
