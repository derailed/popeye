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

func TestNSSanitizer(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)
	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.Namespace](ctx, l.DB, "core/ns/1.yaml", internal.Glossary[internal.NS]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))

	ns := NewNamespace(test.MakeCollector(t), dba)
	assert.Nil(t, ns.Lint(test.MakeContext("v1/namespaces", "ns")))
	assert.Equal(t, 3, len(ns.Outcome()))

	ii := ns.Outcome()["default"]
	assert.Equal(t, 0, len(ii))

	ii = ns.Outcome()["ns1"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, "[POP-400] Used? Unable to locate resource reference", ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = ns.Outcome()["ns2"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, "[POP-800] Namespace is inactive", ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
}
