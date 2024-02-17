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
)

func TestRSLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*appsv1.ReplicaSet](ctx, l.DB, "apps/rs/1.yaml", internal.Glossary[internal.RS]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))

	rs := NewReplicaSet(test.MakeCollector(t), dba)
	assert.Nil(t, rs.Lint(test.MakeContext("apps/v1/replicasets", "replicasets")))
	assert.Equal(t, 2, len(rs.Outcome()))

	ii := rs.Outcome()["default/rs1"]
	assert.Equal(t, 0, len(ii))

	ii = rs.Outcome()["default/rs2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1120] Unhealthy ReplicaSet 2 desired but have 0 ready`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
}
