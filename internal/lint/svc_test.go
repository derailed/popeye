// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/internal/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestSVCLint(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.Service](ctx, l.DB, "core/svc/1.yaml", internal.Glossary[internal.SVC]))
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))
	assert.NoError(t, test.LoadDB[*v1.Endpoints](ctx, l.DB, "core/ep/1.yaml", internal.Glossary[internal.EP]))

	svc := NewService(test.MakeCollector(t), dba)
	assert.Nil(t, svc.Lint(test.MakeContext("v1/pods", "pods")))
	assert.Equal(t, 5, len(svc.Outcome()))

	ii := svc.Outcome()["default/p1"]
	assert.Equal(t, 0, len(ii))

}

func Test_svcCheckEndpoints(t *testing.T) {
	uu := map[string]struct {
		kind     v1.ServiceType
		fqn, key string
		issues   issues.Issues
	}{
		"empty": {
			issues: issues.Issues{
				{
					Group:   "__root__",
					GVR:     "v1/services",
					Level:   rules.ErrorLevel,
					Message: "[POP-1105] No associated endpoints found",
				},
			},
		},
		"external": {
			kind: v1.ServiceTypeExternalName,
		},
		"no-ep": {
			kind: v1.ServiceTypeNodePort,
			fqn:  "default/svc3",
			issues: issues.Issues{
				{
					Group:   "__root__",
					GVR:     "v1/services",
					Level:   rules.ErrorLevel,
					Message: "[POP-1105] No associated endpoints found",
				},
			},
		},
		"nodeport": {
			kind: v1.ServiceTypeNodePort,
			fqn:  "default/svc2",
			issues: issues.Issues{
				{
					Group:   "__root__",
					GVR:     "v1/services",
					Level:   rules.WarnLevel,
					Message: "[POP-1109] Single endpoint is associated with this service",
				},
			},
		},
		"no-subset": {
			kind: v1.ServiceTypeNodePort,
			fqn:  "default/svc4",
			issues: issues.Issues{
				{
					Group:   "__root__",
					GVR:     "v1/services",
					Level:   rules.WarnLevel,
					Message: "[POP-1110] Match EP has no subsets",
				},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			dba, err := test.NewTestDB()
			assert.NoError(t, err)
			l := db.NewLoader(dba)
			ctx := test.MakeContext("v1/services", "services")
			ctx = context.WithValue(ctx, internal.KeyConfig, test.MakeConfig(t))

			assert.NoError(t, test.LoadDB[*v1.Endpoints](ctx, l.DB, "core/ep/1.yaml", internal.Glossary[internal.EP]))

			s := NewService(test.MakeCollector(t), dba)
			if u.fqn != "" {
				ctx = internal.WithSpec(ctx, specFor(u.fqn, nil))
			}
			s.checkEndpoints(ctx, u.fqn, u.kind)

			assert.Equal(t, u.issues, s.Outcome()[u.fqn])
		})
	}
}
