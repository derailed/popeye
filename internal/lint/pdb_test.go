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
	polv1 "k8s.io/api/policy/v1"
)

func TestPDBLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*polv1.PodDisruptionBudget](ctx, l.DB, "pol/pdb/1.yaml", internal.Glossary[internal.PDB]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))

	pdb := NewPodDisruptionBudget(test.MakeCollector(t), dba)
	assert.Nil(t, pdb.Lint(test.MakeContext("policy/v1/poddisruptionbudgets", "poddisruptionbudgets")))
	assert.Equal(t, 5, len(pdb.Outcome()))

	ii := pdb.Outcome()["default/pdb1"]
	assert.Equal(t, 0, len(ii))

	ii = pdb.Outcome()["default/pdb2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-900] No pods match pdb selector: app=p2`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)

	ii = pdb.Outcome()["default/pdb3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-900] No pods match pdb selector: app=test4`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)

	ii = pdb.Outcome()["default/pdb4"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-900] No pods match pdb selector: app=test5`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)

	ii = pdb.Outcome()["default/pdb4-1"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-900] No pods match pdb selector: app=test5`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
}
