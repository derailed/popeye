// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/derailed/popeye/internal/cache"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestCluster(t *testing.T) {
	v, err := semver.NewVersion("1.9")
	assert.NoError(t, err)

	c := cache.NewCluster(v)
	v1, err := c.ListVersion()
	assert.NoError(t, err)
	assert.Equal(t, v, v1)
}
