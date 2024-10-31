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

func TestConfigMapLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.ConfigMap](ctx, l.DB, "core/cm/1.yaml", internal.Glossary[internal.CM]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))

	cm := NewConfigMap(test.MakeCollector(t), dba)
	assert.Nil(t, cm.Lint(test.MakeContext("v1/configmaps", "configmaps")))
	assert.Equal(t, 4, len(cm.Outcome()))

	ii := cm.Outcome()["default/cm1"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, "[POP-401] Key \"fred.yaml\" used? Unable to locate key reference", ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = cm.Outcome()["default/cm2"]
	assert.Equal(t, 0, len(ii))

	ii = cm.Outcome()["default/cm3"]
	assert.Equal(t, 0, len(ii))

	ii = cm.Outcome()["default/cm4"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)
}
