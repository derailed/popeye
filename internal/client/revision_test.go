// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package client_test

import (
	"testing"

	"github.com/derailed/popeye/internal/client"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/version"
)

func TestNewRevision(t *testing.T) {
	uu := map[string]struct {
		info         *version.Info
		major, minor int
	}{
		"plain": {
			info:  &version.Info{Major: "1", Minor: "18+"},
			major: 1,
			minor: 18,
		},
		"no-plus": {
			info:  &version.Info{Major: "1", Minor: "18"},
			major: 1,
			minor: 18,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			r, err := client.NewRevision(u.info)
			assert.Nil(t, err)
			assert.Equal(t, u.major, r.Major)
			assert.Equal(t, u.minor, r.Minor)
		})
	}
}
