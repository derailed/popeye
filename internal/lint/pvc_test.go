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

func TestPVCLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.PersistentVolumeClaim](ctx, l.DB, "core/pvc/1.yaml", internal.Glossary[internal.PVC]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))

	pvc := NewPersistentVolumeClaim(test.MakeCollector(t), dba)
	assert.Nil(t, pvc.Lint(test.MakeContext("v1/persistentvolumeclaims", "persistentvolumeclaims")))
	assert.Equal(t, 3, len(pvc.Outcome()))

	ii := pvc.Outcome()["default/pvc1"]
	assert.Equal(t, 0, len(ii))

	ii = pvc.Outcome()["default/pvc2"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-1004] Lost claim detected`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[1].Message)
	assert.Equal(t, rules.InfoLevel, ii[1].Level)

	ii = pvc.Outcome()["default/pvc3"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-1003] Pending claim detected`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[1].Message)
	assert.Equal(t, rules.InfoLevel, ii[1].Level)
}
