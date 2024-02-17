// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"sync"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/test"
	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
)

func TestIngressRefs(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*netv1.Ingress](ctx, l.DB, "net/ingress/1.yaml", internal.Glossary[internal.ING]))

	var refs sync.Map
	ing := NewIngress(dba)
	assert.NoError(t, ing.IngressRefs(&refs))

	_, ok := refs.Load("sec:default/foo")
	assert.Equal(t, ok, true)
}
