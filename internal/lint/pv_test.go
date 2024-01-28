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
)

func TestPVLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.PersistentVolume](ctx, l.DB, "core/pv/1.yaml", internal.Glossary[internal.PV]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))

	pv := NewPersistentVolume(test.MakeCollector(t), dba)
	assert.Nil(t, pv.Lint(test.MakeContext("v1/persistentvolumes", "persistentvolumes")))
	assert.Equal(t, 4, len(pv.Outcome()))

	ii := pv.Outcome()["default/pv1"]
	assert.Equal(t, 0, len(ii))

	ii = pv.Outcome()["default/pv2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1002] Lost volume detected`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)

	ii = pv.Outcome()["default/pv3"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1000] Available volume detected`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = pv.Outcome()["default/pv4"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1001] Pending volume detected`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
}
