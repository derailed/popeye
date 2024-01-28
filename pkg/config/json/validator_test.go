// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package json_test

import (
	"os"
	"testing"

	"github.com/derailed/popeye/pkg/config/json"
	"github.com/stretchr/testify/assert"
)

func TestValidateSpinach(t *testing.T) {
	uu := map[string]struct {
		f   string
		err string
	}{
		"happy": {
			f: "testdata/1.yaml",
		},
		"toast": {
			f:   "testdata/toast.yaml",
			err: "Additional property rbac.authorization.k8s.io/v1/clusterroles is not allowed",
		},
	}

	v := json.NewValidator()
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			bb, err := os.ReadFile(u.f)
			assert.NoError(t, err)
			if err := v.Validate(json.SpinachSchema, bb); err != nil {
				assert.Equal(t, u.err, err.Error())
			}
		})
	}
}
