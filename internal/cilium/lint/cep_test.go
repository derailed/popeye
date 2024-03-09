// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"testing"

	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cilium"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/internal/test"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestCiliumEndpoint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v2.CiliumEndpoint](ctx, l.DB, "cep/1.yaml", internal.Glossary[cilium.CEP]))
	assert.NoError(t, test.LoadDB[*v2.CiliumIdentity](ctx, l.DB, "cid/1.yaml", internal.Glossary[cilium.CID]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "../../../lint/testdata/core/pod/1.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*v1.Node](ctx, l.DB, "../../../lint/testdata/core/node/1.yaml", internal.Glossary[internal.NO]))
	assert.NoError(t, test.LoadDB[*v1.Namespace](ctx, l.DB, "../../../lint/testdata/core/ns/1.yaml", internal.Glossary[internal.NS]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "../../../lint/testdata/core/sa/1.yaml", internal.Glossary[internal.SA]))

	li := NewCiliumEndpoint(test.MakeCollector(t), dba)
	assert.Nil(t, li.Lint(test.MakeContext("cilium.io/v2/ciliumendpoints", "ciliumendpoints")))
	assert.Equal(t, 2, len(li.Outcome()))

	ii := li.Outcome()["default/cep1"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1703] Pod owner is not in a running state: default/p1 ()`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)

	ii = li.Outcome()["default/cep2"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-1704] References an unknown owner ref: "default/p2"`, ii[0].Message)
	assert.Equal(t, rules.ErrorLevel, ii[0].Level)
	assert.Equal(t, `[POP-1702] References an unknown node IP: "172.19.0.2"`, ii[1].Message)
	assert.Equal(t, rules.ErrorLevel, ii[1].Level)
}
