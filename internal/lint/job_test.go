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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestJSpecFor(t *testing.T) {
	tests := map[string]struct {
		fqn  string
		job  *batchv1.Job
		want rules.Spec
	}{
		"full": {
			fqn: "default/p1",
			job: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"p1": "blee"},
					Annotations: map[string]string{"default": "fred"},
				},
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							InitContainers: []v1.Container{
								{Name: "ic1"},
							},
							Containers: []v1.Container{
								{Name: "c1"},
								{Name: "c2"},
							},
						},
					},
				},
			},
			want: rules.Spec{
				FQN:         "default/p1",
				Labels:      rules.Labels{"p1": "blee"},
				Annotations: rules.Labels{"default": "fred"},
				Containers:  []string{"ic1", "c1", "c2"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := JSpecFor(tc.fqn, tc.job)
			assert.Equal(t, tc.want.FQN, got.FQN)
			assert.Equal(t, tc.want.Labels, got.Labels)
			assert.Equal(t, tc.want.Annotations, got.Annotations)
			assert.ElementsMatch(t, tc.want.Containers, got.Containers)
		})
	}
}
