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

func TestJobLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*batchv1.Job](ctx, l.DB, "batch/job/1.yaml", internal.Glossary[internal.JOB]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*mv1beta1.PodMetrics](ctx, l.DB, "mx/pod/1.yaml", internal.Glossary[internal.PMX]))

	j := NewJob(test.MakeCollector(t), dba)
	assert.Nil(t, j.Lint(test.MakeContext("batch/v1/jobs", "jobs")))
	assert.Equal(t, 3, len(j.Outcome()))

	ii := j.Outcome()["default/j1"]
	assert.Equal(t, 0, len(ii))

	ii = j.Outcome()["default/j2"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-100] Untagged docker image in use`, ii[0].Message)
	assert.Equal(t, `[POP-106] No resources requests/limits defined`, ii[1].Message)
	assert.Equal(t, rules.WarnLevel, ii[1].Level)
}
