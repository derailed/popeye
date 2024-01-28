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
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestHPALint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*autoscalingv1.HorizontalPodAutoscaler](ctx, l.DB, "autoscaling/hpa/1.yaml", internal.Glossary[internal.HPA]))
	assert.NoError(t, test.LoadDB[*appsv1.Deployment](ctx, l.DB, "apps/dp/1.yaml", internal.Glossary[internal.DP]))
	assert.NoError(t, test.LoadDB[*appsv1.ReplicaSet](ctx, l.DB, "apps/rs/1.yaml", internal.Glossary[internal.RS]))
	assert.NoError(t, test.LoadDB[*appsv1.StatefulSet](ctx, l.DB, "apps/sts/1.yaml", internal.Glossary[internal.STS]))
	assert.NoError(t, test.LoadDB[*v1.Node](ctx, l.DB, "core/node/1.yaml", internal.Glossary[internal.NO]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/2.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*mv1beta1.PodMetrics](ctx, l.DB, "mx/pod/1.yaml", internal.Glossary[internal.PMX]))
	assert.NoError(t, test.LoadDB[*mv1beta1.NodeMetrics](ctx, l.DB, "mx/node/1.yaml", internal.Glossary[internal.NMX]))

	hpa := NewHorizontalPodAutoscaler(test.MakeCollector(t), dba)
	assert.Nil(t, hpa.Lint(test.MakeContext("autoscaling/v1/horizontalpodautoscalers", "horizontalpodautoscalers")))
	assert.Equal(t, 7, len(hpa.Outcome()))

	ii := hpa.Outcome()["default/hpa1"]
	assert.Equal(t, 1, len(ii))

	ii = hpa.Outcome()["default/hpa2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-600] HPA default/hpa2 references a deployment which does not exist: default/dp-toast`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)

	ii = hpa.Outcome()["default/hpa3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-600] HPA default/hpa3 references a replicaset which does not exist: default/rs-toast`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)

	ii = hpa.Outcome()["default/hpa4"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-600] HPA default/hpa4 references a statefulset which does not exist: default/sts-toast`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)

	ii = hpa.Outcome()["default/hpa5"]
	assert.Equal(t, 1, len(ii))

	ii = hpa.Outcome()["default/hpa6"]
	assert.Equal(t, 1, len(ii))

}
