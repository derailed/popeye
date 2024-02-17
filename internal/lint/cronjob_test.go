// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/internal/test"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestCronJobLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*batchv1.CronJob](ctx, l.DB, "batch/cjob/1.yaml", internal.Glossary[internal.CJOB]))
	assert.NoError(t, test.LoadDB[*batchv1.Job](ctx, l.DB, "batch/job/1.yaml", internal.Glossary[internal.JOB]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*mv1beta1.PodMetrics](ctx, l.DB, "mx/pod/1.yaml", internal.Glossary[internal.PMX]))

	cj := NewCronJob(test.MakeCollector(t), dba)
	assert.Nil(t, cj.Lint(test.MakeContext("batch/v1/cronjobs", "cronjobs")))
	assert.Equal(t, 2, len(cj.Outcome()))

	ii := cj.Outcome()["default/cj1"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-503] At current load, CPU under allocated. Current:2000m vs Requested:1m (200000.00%)`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-505] At current load, Memory under allocated. Current:20Mi vs Requested:1Mi (2000.00%)`, ii[1].Message)
	assert.Equal(t, rules.WarnLevel, ii[1].Level)

	ii = cj.Outcome()["default/cj2"]
	assert.Equal(t, 6, len(ii))
	assert.Equal(t, `[POP-1500] CronJob is suspended`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-1501] No active jobs detected`, ii[1].Message)
	assert.Equal(t, rules.InfoLevel, ii[1].Level)
	assert.Equal(t, `[POP-1502] CronJob has not run yet or is failing`, ii[2].Message)
	assert.Equal(t, rules.WarnLevel, ii[2].Level)
	assert.Equal(t, `[POP-307] CronJob references a non existing ServiceAccount: "sa-bozo"`, ii[3].Message)
	assert.Equal(t, rules.WarnLevel, ii[3].Level)
	assert.Equal(t, `[POP-100] Untagged docker image in use`, ii[4].Message)
	assert.Equal(t, rules.ErrorLevel, ii[4].Level)
	assert.Equal(t, `[POP-106] No resources requests/limits defined`, ii[5].Message)
	assert.Equal(t, rules.WarnLevel, ii[5].Level)
}
