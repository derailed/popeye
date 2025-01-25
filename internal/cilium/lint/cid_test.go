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
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestCiliumIdentity(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v2.CiliumIdentity](ctx, l.DB, "cid/1.yaml", internal.Glossary[cilium.CID]))
	assert.NoError(t, test.LoadDB[*v2.CiliumEndpoint](ctx, l.DB, "cep/1.yaml", internal.Glossary[cilium.CEP]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "../../../lint/testdata/core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*v1.Namespace](ctx, l.DB, "../../../lint/testdata/core/ns/1.yaml", internal.Glossary[internal.NS]))

	li := NewCiliumIdentity(test.MakeCollector(t), dba)
	assert.Nil(t, li.Lint(test.MakeContext("cilium.io/v2/ciliumidentities", "ciliumidentities")))
	assert.Equal(t, 3, len(li.Outcome()))

	ii := li.Outcome()["default/100"]
	assert.Equal(t, 0, len(ii))

	ii = li.Outcome()["ns1/200"]
	assert.Equal(t, 3, len(ii))
	assert.Equal(t, "[POP-1600] Stale? unable to locate matching Cilium Endpoint", ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-1604] Namespace mismatch with security labels namespace: "ns1" vs "ns2"`, ii[1].Message)
	assert.Equal(t, rules.WarnLevel, ii[1].Level)
	assert.Equal(t, `[POP-307] CiliumIdentity references a non existing ServiceAccount: "ns1/sa1"`, ii[2].Message)
	assert.Equal(t, rules.WarnLevel, ii[2].Level)

	ii = li.Outcome()["default/300"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1603] Missing security namespace label: "k8s:io.kubernetes.pod.namespace"`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
}
