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
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNodeSanitizer(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.Node](ctx, l.DB, "core/node/1.yaml", internal.Glossary[internal.NO]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*mv1beta1.NodeMetrics](ctx, l.DB, "mx/node/1.yaml", internal.Glossary[internal.NMX]))

	no := NewNode(test.MakeCollector(t), dba)
	assert.Nil(t, no.Lint(test.MakeContext("v1/nodes", "nodes")))
	assert.Equal(t, 5, len(no.Outcome()))

	ii := no.Outcome()["n1"]
	assert.Equal(t, 0, len(ii))

	ii = no.Outcome()["n2"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-707] No network configured on node`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-700] Found taint "t2" but no pod can tolerate`, ii[1].Message)
	assert.Equal(t, rules.WarnLevel, ii[1].Level)

	ii = no.Outcome()["n3"]
	assert.Equal(t, 5, len(ii))
	assert.Equal(t, `[POP-704] Insufficient memory`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-705] Insufficient disk space`, ii[1].Message)
	assert.Equal(t, rules.WarnLevel, ii[1].Level)
	assert.Equal(t, `[POP-706] Insufficient PIDs on Node`, ii[2].Message)
	assert.Equal(t, rules.ErrorLevel, ii[2].Level)
	assert.Equal(t, `[POP-707] No network configured on node`, ii[3].Message)
	assert.Equal(t, rules.ErrorLevel, ii[3].Level)
	assert.Equal(t, `[POP-708] No node metrics available`, ii[4].Message)
	assert.Equal(t, rules.InfoLevel, ii[4].Level)

	ii = no.Outcome()["n4"]
	assert.Equal(t, 4, len(ii))
	assert.Equal(t, `[POP-711] Scheduling disabled`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-701] Node has an unknown condition`, ii[1].Message)
	assert.Equal(t, rules.WarnLevel, ii[1].Level)
	assert.Equal(t, `[POP-702] Node is not in ready state`, ii[2].Message)
	assert.Equal(t, rules.ErrorLevel, ii[2].Level)
	assert.Equal(t, `[POP-708] No node metrics available`, ii[3].Message)
	assert.Equal(t, rules.InfoLevel, ii[3].Level)

	ii = no.Outcome()["n5"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-709] CPU threshold (80%) reached 20000%`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-710] Memory threshold (80%) reached 400%`, ii[1].Message)
	assert.Equal(t, rules.WarnLevel, ii[1].Level)
}
