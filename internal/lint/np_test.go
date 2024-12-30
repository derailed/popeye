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
	netv1 "k8s.io/api/networking/v1"
)

func TestNPLintDenyAll(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*netv1.NetworkPolicy](ctx, l.DB, "net/np/2.yaml", internal.Glossary[internal.NP]))
	assert.NoError(t, test.LoadDB[*v1.Namespace](ctx, l.DB, "core/ns/1.yaml", internal.Glossary[internal.NS]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))

	np := NewNetworkPolicy(test.MakeCollector(t), dba)
	assert.Nil(t, np.Lint(test.MakeContext("networking.k8s.io/v1/networkpolicies", "networkpolicies")))
	assert.Equal(t, 8, len(np.Outcome()))

	ii := np.Outcome()["default/deny-all"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1203] Deny all policy in effect`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = np.Outcome()["default/deny-all-ing"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1203] Deny all ingress policy in effect`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = np.Outcome()["default/deny-all-eg"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1203] Deny all egress policy in effect`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = np.Outcome()["default/allow-all"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1203] Allow all policy in effect`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = np.Outcome()["default/allow-all-ing"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1203] Allow all ingress policy in effect`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = np.Outcome()["default/allow-all-eg"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1203] Allow all egress policy in effect`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = np.Outcome()["default/ip-block-all-ing"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-1206] No pods matched egress IPBlock 172.2.0.0/24`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-1203] Deny all ingress policy in effect`, ii[1].Message)
	assert.Equal(t, rules.InfoLevel, ii[1].Level)

	ii = np.Outcome()["default/ip-block-all-eg"]
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, `[POP-1206] No pods matched ingress IPBlock 172.2.0.0/24`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-1203] Deny all egress policy in effect`, ii[1].Message)
	assert.Equal(t, rules.InfoLevel, ii[1].Level)
}

func TestNPLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*netv1.NetworkPolicy](ctx, l.DB, "net/np/1.yaml", internal.Glossary[internal.NP]))
	assert.NoError(t, test.LoadDB[*v1.Namespace](ctx, l.DB, "core/ns/1.yaml", internal.Glossary[internal.NS]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))

	np := NewNetworkPolicy(test.MakeCollector(t), dba)
	assert.Nil(t, np.Lint(test.MakeContext("networking.k8s.io/v1/networkpolicies", "networkpolicies")))
	assert.Equal(t, 3, len(np.Outcome()))

	ii := np.Outcome()["default/np1"]
	assert.Equal(t, 0, len(ii))

	ii = np.Outcome()["default/np2"]
	assert.Equal(t, 3, len(ii))
	assert.Equal(t, `[POP-1207] No pods matched except ingress IPBlock 172.1.1.0/24`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-1208] No pods match ingress pod selector: app=p2 in namespace: ns2`, ii[1].Message)
	assert.Equal(t, rules.WarnLevel, ii[1].Level)
	assert.Equal(t, `[POP-1206] No pods matched egress IPBlock 172.0.0.0/24`, ii[2].Message)
	assert.Equal(t, rules.WarnLevel, ii[2].Level)

	ii = np.Outcome()["default/np3"]
	assert.Equal(t, 6, len(ii))
	assert.Equal(t, `[POP-1200] No pods match pod selector: app=p-bozo`, ii[0].Message)
	assert.Equal(t, rules.WarnLevel, ii[0].Level)
	assert.Equal(t, `[POP-1206] No pods matched ingress IPBlock 172.2.0.0/16`, ii[1].Message)
	assert.Equal(t, rules.WarnLevel, ii[1].Level)
	assert.Equal(t, `[POP-1207] No pods matched except ingress IPBlock 172.2.1.0/24`, ii[2].Message)
	assert.Equal(t, rules.WarnLevel, ii[2].Level)
	assert.Equal(t, `[POP-1201] No namespaces match ingress namespace selector: app-In-ns-bozo`, ii[3].Message)
	assert.Equal(t, rules.WarnLevel, ii[3].Level)
	assert.Equal(t, `[POP-1202] No pods match ingress pod selector: app=pod-bozo`, ii[4].Message)
	assert.Equal(t, rules.WarnLevel, ii[4].Level)
	assert.Equal(t, `[POP-1208] No pods match egress pod selector: app=p1-missing in namespace: default`, ii[5].Message)
	assert.Equal(t, rules.WarnLevel, ii[5].Level)
}

func Test_npDefaultDenyAll(t *testing.T) {
	uu := map[string]struct {
		path string
		e    bool
	}{
		"open": {
			path: "net/np/a.yaml",
		},
		"deny-all": {
			path: "net/np/deny-all.yaml",
			e:    true,
		},
		"allow-all-ing": {
			path: "net/np/allow-all-ing.yaml",
		},
		"no-selector": {
			path: "net/np/d.yaml",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			np, err := test.LoadRes[netv1.NetworkPolicy](u.path)
			assert.NoError(t, err)
			assert.Equal(t, u.e, isDefaultDenyAll(np))
		})
	}
}
