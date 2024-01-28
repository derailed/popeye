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
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, "[POP-400] Used? Unable to locate resource reference", ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)

	ii = cm.Outcome()["default/cm3"]
	assert.Equal(t, 0, len(ii))

	ii = cm.Outcome()["default/cm4"]
	assert.Equal(t, 1, len(ii))
	assert.Equal(t, `[POP-400] Used? Unable to locate resource reference`, ii[0].Message)
	assert.Equal(t, rules.InfoLevel, ii[0].Level)
}

// ----------------------------------------------------------------------------
// Helpers...

// type mockConfigMap struct{}

// func newMockConfigMap() mockConfigMap {
// 	return mockConfigMap{}
// }

// func (c mockConfigMap) PodRefs(refs *sync.Map) {
// 	refs.Store("cm:default/cm1", internal.StringSet{
// 		"k1": internal.Blank,
// 		"k2": internal.Blank,
// 	})
// 	refs.Store("cm:default/cm2", internal.AllKeys)
// 	refs.Store("cm:default/cm4", internal.StringSet{
// 		"k1": internal.Blank,
// 	})
// }

// func (c mockConfigMap) ListConfigMaps() map[string]*v1.ConfigMap {
// 	return map[string]*v1.ConfigMap{
// 		"default/cm1": makeMockConfigMap("cm1"),
// 		"default/cm2": makeMockConfigMap("cm2"),
// 		"default/cm3": makeMockConfigMap("cm3"),
// 		"default/cm4": makeMockConfigMap("cm4"),
// 	}
// }

// func makeMockConfigMap(n string) *v1.ConfigMap {
// 	return &v1.ConfigMap{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      n,
// 			Namespace: "default",
// 		},
// 		Data: map[string]string{
// 			"k1": "",
// 			"k2": "",
// 		},
// 	}
// }
