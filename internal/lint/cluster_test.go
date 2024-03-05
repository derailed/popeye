// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/version"

	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/internal/test"
)

func TestClusterLint(t *testing.T) {
	uu := map[string]struct {
		major, minor string
		metrics      bool
		e            issues.Outcome
	}{
		"good": {
			major: "1", minor: "29",
			metrics: true,
			e: map[string]issues.Issues{
				"Version": {
					{
						GVR:     "clusters",
						Group:   issues.Root,
						Message: "[POP-406] K8s version OK",
						Level:   rules.OkLevel,
					},
				},
			},
		},
		"plus": {
			major: "1", minor: "29+",
			metrics: true,
			e: map[string]issues.Issues{
				"Version": {
					{
						GVR:     "clusters",
						Group:   issues.Root,
						Message: "[POP-406] K8s version OK",
						Level:   rules.OkLevel,
					},
				},
			},
		},
		"gizzard": {
			major: "1", minor: "11",
			metrics: false,
			e: map[string]issues.Issues{
				"Version": {
					{
						GVR:     "clusters",
						Group:   issues.Root,
						Message: "[POP-405] Is this a jurassic cluster? Might want to upgrade K8s a bit",
						Level:   rules.WarnLevel,
					},
				},
			},
		},
	}

	ctx := test.MakeContext("clusters", "cluster")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cl := NewCluster(
				test.MakeCollector(t),
				newMockCluster(u.major, u.minor, u.metrics),
			)

			assert.Nil(t, cl.Lint(ctx))
			assert.Equal(t, u.e, cl.Outcome())
		})
	}
}

// Helpers...

type mockCluster struct {
	major, minor string
	metrics      bool
}

func newMockCluster(major, minor string, metrics bool) mockCluster {
	return mockCluster{major: major, minor: minor, metrics: metrics}
}

func (c mockCluster) ListVersion() (*semver.Version, error) {
	return dag.ParseVersion(&version.Info{Major: c.major, Minor: c.minor})
}

func (c mockCluster) HasMetrics() bool {
	return c.metrics
}
