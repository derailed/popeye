// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/test"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/pkg/config/json"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	polv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestPodNPLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/3.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/2.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*v1.Namespace](ctx, l.DB, "core/ns/1.yaml", internal.Glossary[internal.NS]))
	assert.NoError(t, test.LoadDB[*polv1.PodDisruptionBudget](ctx, l.DB, "pol/pdb/1.yaml", internal.Glossary[internal.PDB]))
	assert.NoError(t, test.LoadDB[*netv1.NetworkPolicy](ctx, l.DB, "net/np/3.yaml", internal.Glossary[internal.NP]))
	assert.NoError(t, test.LoadDB[*mv1beta1.PodMetrics](ctx, l.DB, "mx/pod/1.yaml", internal.Glossary[internal.PMX]))

	po := NewPod(test.MakeCollector(t), dba)
	assert.Nil(t, po.Lint(test.MakeContext("v1/pods", "pods")))
	assert.Equal(t, 2, len(po.Outcome()))

	ii := po.Outcome()["ns1/p1"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-1204] Pod egress is not secured by a network policy`, ii[0].Message)

	ii = po.Outcome()["ns2/p2"]
	assert.Equal(t, 0, len(ii))
}

func TestPodCheckSecure(t *testing.T) {
	uu := map[string]struct {
		pod    v1.Pod
		issues int
	}{
		"cool_1": {
			pod:    makeSecPod(secNonRootSet, secNonRootSet, secNonRootSet, secNonRootSet),
			issues: 1,
		},
		"cool_2": {
			pod:    makeSecPod(secNonRootSet, secNonRootUnset, secNonRootUnset, secNonRootUnset),
			issues: 1,
		},
		"cool_3": {
			pod:    makeSecPod(secNonRootUnset, secNonRootSet, secNonRootSet, secNonRootSet),
			issues: 1,
		},
		"cool_4": {
			pod:    makeSecPod(secNonRootUndefined, secNonRootSet, secNonRootSet, secNonRootSet),
			issues: 1,
		},
		"cool_5": {
			pod:    makeSecPod(secNonRootSet, secNonRootUndefined, secNonRootUndefined, secNonRootUndefined),
			issues: 1,
		},
		"hacked_1": {
			pod:    makeSecPod(secNonRootUndefined, secNonRootUndefined, secNonRootUndefined, secNonRootUndefined),
			issues: 5,
		},
		"hacked_2": {
			pod:    makeSecPod(secNonRootUndefined, secNonRootUnset, secNonRootUndefined, secNonRootUndefined),
			issues: 5,
		},
		"hacked_3": {
			pod:    makeSecPod(secNonRootUndefined, secNonRootSet, secNonRootUndefined, secNonRootUndefined),
			issues: 4,
		},
		"hacked_4": {
			pod:    makeSecPod(secNonRootUndefined, secNonRootUnset, secNonRootSet, secNonRootUndefined),
			issues: 4,
		},
		"toast": {
			pod:    makeSecPod(secNonRootUndefined, secNonRootUndefined, secNonRootUndefined, secNonRootUndefined),
			issues: 5,
		},
	}

	ctx := test.MakeContext("v1/pods", "po")
	ctx = internal.WithSpec(ctx, SpecFor("default/p1", nil))
	ctx = context.WithValue(ctx, internal.KeyConfig, test.MakeConfig(t))
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/2.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*polv1.PodDisruptionBudget](ctx, l.DB, "pol/pdb/1.yaml", internal.Glossary[internal.PDB]))
	assert.NoError(t, test.LoadDB[*mv1beta1.PodMetrics](ctx, l.DB, "mx/pod/1.yaml", internal.Glossary[internal.PMX]))

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := NewPod(test.MakeCollector(t), dba)
			p.checkSecure(ctx, "default/p1", u.pod.Spec, true)
			assert.Equal(t, u.issues, len(p.Outcome()["default/p1"]))
		})
	}
}

func TestPodLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/2.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*polv1.PodDisruptionBudget](ctx, l.DB, "pol/pdb/1.yaml", internal.Glossary[internal.PDB]))
	assert.NoError(t, test.LoadDB[*netv1.NetworkPolicy](ctx, l.DB, "net/np/1.yaml", internal.Glossary[internal.NP]))
	assert.NoError(t, test.LoadDB[*mv1beta1.PodMetrics](ctx, l.DB, "mx/pod/1.yaml", internal.Glossary[internal.PMX]))

	po := NewPod(test.MakeCollector(t), dba)
	po.Collector.Config.Registries = []string{"dorker.io"}
	assert.Nil(t, po.Lint(test.MakeContext("v1/pods", "pods")))
	assert.Equal(t, 5, len(po.Outcome()))

	ii := po.Outcome()["default/p1"]
	assert.Equal(t, 0, len(ii))

	ii = po.Outcome()["default/p2"]
	assert.Equal(t, 6, len(ii))
	assert.Equal(t, `[POP-207] Pod is in an unhappy phase ()`, ii[0].Message)
	assert.Equal(t, `[POP-208] Unmanaged pod detected. Best to use a controller`, ii[1].Message)
	assert.Equal(t, `[POP-1204] Pod ingress is not secured by a network policy`, ii[2].Message)
	assert.Equal(t, `[POP-1204] Pod egress is not secured by a network policy`, ii[3].Message)
	assert.Equal(t, `[POP-206] Pod has no associated PodDisruptionBudget`, ii[4].Message)
	assert.Equal(t, `[POP-301] Connects to API Server? ServiceAccount token is mounted`, ii[5].Message)

	ii = po.Outcome()["default/p3"]
	assert.Equal(t, 6, len(ii))
	assert.Equal(t, `[POP-105] Liveness uses a port#, prefer a named port`, ii[0].Message)
	assert.Equal(t, `[POP-105] Readiness uses a port#, prefer a named port`, ii[1].Message)
	assert.Equal(t, `[POP-1204] Pod ingress is not secured by a network policy`, ii[2].Message)
	assert.Equal(t, `[POP-1204] Pod egress is not secured by a network policy`, ii[3].Message)
	assert.Equal(t, `[POP-301] Connects to API Server? ServiceAccount token is mounted`, ii[4].Message)
	assert.Equal(t, `[POP-109] CPU Current/Request (2000m/1000m) reached user 80% threshold (200%)`, ii[5].Message)

	ii = po.Outcome()["default/p4"]
	assert.Equal(t, 15, len(ii))
	assert.Equal(t, `[POP-204] Pod is not ready [0/1]`, ii[0].Message)
	assert.Equal(t, `[POP-204] Pod is not ready [0/2]`, ii[1].Message)
	assert.Equal(t, `[POP-100] Untagged docker image in use`, ii[2].Message)
	assert.Equal(t, `[POP-113] Container image "zorg" is not hosted on an allowed docker registry`, ii[3].Message)
	assert.Equal(t, `[POP-106] No resources requests/limits defined`, ii[4].Message)
	assert.Equal(t, `[POP-100] Untagged docker image in use`, ii[5].Message)
	assert.Equal(t, `[POP-113] Container image "blee" is not hosted on an allowed docker registry`, ii[6].Message)
	assert.Equal(t, `[POP-101] Image tagged "latest" in use`, ii[7].Message)
	assert.Equal(t, `[POP-113] Container image "zorg:latest" is not hosted on an allowed docker registry`, ii[8].Message)
	assert.Equal(t, `[POP-107] No resource limits defined`, ii[9].Message)
	assert.Equal(t, `[POP-208] Unmanaged pod detected. Best to use a controller`, ii[10].Message)
	assert.Equal(t, `[POP-1204] Pod ingress is not secured by a network policy`, ii[11].Message)
	assert.Equal(t, `[POP-1204] Pod egress is not secured by a network policy`, ii[12].Message)
	assert.Equal(t, `[POP-300] Uses "default" ServiceAccount`, ii[13].Message)
	assert.Equal(t, `[POP-301] Connects to API Server? ServiceAccount token is mounted`, ii[14].Message)

	ii = po.Outcome()["default/p5"]
	assert.Equal(t, 7, len(ii))
	assert.Equal(t, `[POP-113] Container image "blee:v1.2" is not hosted on an allowed docker registry`, ii[0].Message)
	assert.Equal(t, `[POP-106] No resources requests/limits defined`, ii[1].Message)
	assert.Equal(t, `[POP-102] No probes defined`, ii[2].Message)
	assert.Equal(t, `[POP-1204] Pod ingress is not secured by a network policy`, ii[3].Message)
	assert.Equal(t, `[POP-1204] Pod egress is not secured by a network policy`, ii[4].Message)
	assert.Equal(t, `[POP-209] Pod is managed by multiple PodDisruptionBudgets (pdb4, pdb4-1)`, ii[5].Message)
	assert.Equal(t, `[POP-301] Connects to API Server? ServiceAccount token is mounted`, ii[6].Message)
}

func TestPodLintExcludes(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/2.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))
	assert.NoError(t, test.LoadDB[*polv1.PodDisruptionBudget](ctx, l.DB, "pol/pdb/1.yaml", internal.Glossary[internal.PDB]))
	assert.NoError(t, test.LoadDB[*netv1.NetworkPolicy](ctx, l.DB, "net/np/1.yaml", internal.Glossary[internal.NP]))
	assert.NoError(t, test.LoadDB[*mv1beta1.PodMetrics](ctx, l.DB, "mx/pod/1.yaml", internal.Glossary[internal.PMX]))

	bb, err := os.ReadFile(filepath.Join("testdata", "config", "1.yaml"))
	assert.NoError(t, err)
	assert.NoError(t, json.NewValidator().Validate(json.SpinachSchema, bb))
	var cfg config.Config
	assert.NoError(t, yaml.Unmarshal(bb, &cfg))

	codes, err := issues.LoadCodes()
	assert.NoError(t, err)
	cc := issues.NewCollector(codes, &cfg)
	po := NewPod(cc, dba)
	po.Collector.Config.Registries = []string{"dorker.io"}
	assert.Nil(t, po.Lint(test.MakeContext("v1/pods", "pods")))
	assert.Equal(t, 5, len(po.Outcome()))

	ii := po.Outcome()["default/p4"]

	assert.Equal(t, 7, len(ii))
	assert.Equal(t, `[POP-101] Image tagged "latest" in use`, ii[0].Message)
	assert.Equal(t, `[POP-107] No resource limits defined`, ii[1].Message)
	assert.Equal(t, `[POP-208] Unmanaged pod detected. Best to use a controller`, ii[2].Message)
	assert.Equal(t, `[POP-1204] Pod ingress is not secured by a network policy`, ii[3].Message)
	assert.Equal(t, `[POP-1204] Pod egress is not secured by a network policy`, ii[4].Message)
	assert.Equal(t, `[POP-300] Uses "default" ServiceAccount`, ii[5].Message)
	assert.Equal(t, `[POP-301] Connects to API Server? ServiceAccount token is mounted`, ii[6].Message)
}

// ----------------------------------------------------------------------------
// Helpers...

type nonRootUser int

const (
	secNonRootUndefined nonRootUser = iota - 1
	secNonRootUnset                 = 0
	secNonRootSet                   = 1
)

func makeSecCO(name string, level nonRootUser) v1.Container {
	t, f := true, false
	var secCtx v1.SecurityContext
	switch level {
	case secNonRootUnset:
		secCtx.RunAsNonRoot = &f
	case secNonRootSet:
		secCtx.RunAsNonRoot = &t
	default:
		secCtx.RunAsNonRoot = nil
	}

	return v1.Container{Name: name, SecurityContext: &secCtx}
}

func makeSecPod(pod, init, co1, co2 nonRootUser) v1.Pod {
	t, f := true, false
	var zero int64
	var secCtx v1.PodSecurityContext
	switch pod {
	case secNonRootUnset:
		secCtx.RunAsNonRoot = &f
	case secNonRootSet:
		secCtx.RunAsNonRoot = &t
	default:
		secCtx.RunAsNonRoot = nil
		secCtx.RunAsUser = &zero
	}

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "p1",
		},
		Spec: v1.PodSpec{
			ServiceAccountName:           "default",
			AutomountServiceAccountToken: &f,
			InitContainers: []v1.Container{
				makeSecCO("ic1", init),
			},
			Containers: []v1.Container{
				makeSecCO("c1", co1),
				makeSecCO("c2", co2),
			},
			SecurityContext: &secCtx,
		},
	}
}
