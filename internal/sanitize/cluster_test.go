// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package sanitize

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
)

func TestClusterSanitize(t *testing.T) {
	uu := map[string]struct {
		major, minor string
		metrics      bool
		e            issues.Outcome
	}{
		"good": {
			major: "1", minor: "15",
			metrics: true,
			e: map[string]issues.Issues{
				"Version": {
					{
						GVR:     "clusters",
						Group:   issues.Root,
						Message: "[POP-406] K8s version OK",
						Level:   config.OkLevel,
					},
				},
			},
		},
		"guizard": {
			major: "1", minor: "11",
			metrics: false,
			e: map[string]issues.Issues{
				"Version": {
					{
						GVR:     "clusters",
						Group:   issues.Root,
						Message: "[POP-405] Is this a jurassic cluster? Might want to upgrade K8s a bit",
						Level:   config.WarnLevel,
					},
				},
				"Metrics": {
					{
						GVR:     "clusters",
						Group:   issues.Root,
						Message: "[POP-402] No metrics-server detected",
						Level:   config.InfoLevel,
					},
				},
			},
		},
	}

	ctx := makeContext("clusters", "cluster")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cl := NewCluster(issues.NewCollector(loadCodes(t), makeConfig(t)), newCluster(u.major, u.minor, u.metrics))

			assert.Nil(t, cl.Sanitize(ctx))
			assert.Equal(t, u.e, cl.Outcome())
		})
	}
}

// Helpers...

func makeConfig(t *testing.T) *config.Config {
	c, err := config.NewConfig(config.NewFlags())
	assert.Nil(t, err)
	return c
}

func makeContext(gvr, section string) context.Context {
	return context.WithValue(context.Background(), internal.KeyRunInfo, internal.RunInfo{
		Section:    section,
		SectionGVR: client.NewGVR(gvr),
	})
}

type cluster struct {
	major, minor string
	metrics      bool
}

func newCluster(major, minor string, metrics bool) cluster {
	return cluster{major: major, minor: minor, metrics: metrics}
}

func (c cluster) ListVersion() (string, string) {
	return c.major, c.minor
}

func (c cluster) HasMetrics() bool {
	return c.metrics
}
