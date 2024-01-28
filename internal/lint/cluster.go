// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/Masterminds/semver"
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
)

const (
	tolerableMajor = 1
	tolerableMinor = 21
)

type (
	// Cluster tracks cluster sanitization.
	Cluster struct {
		*issues.Collector
		ClusterLister
	}

	// ClusterLister list available Clusters on a cluster.
	ClusterLister interface {
		ListVersion() *semver.Version
		HasMetrics() bool
	}
)

// NewCluster returns a new instance.
func NewCluster(co *issues.Collector, lister ClusterLister) *Cluster {
	return &Cluster{
		Collector:     co,
		ClusterLister: lister,
	}
}

// Lint cleanse the resource.
func (c *Cluster) Lint(ctx context.Context) error {
	return c.checkVersion(ctx)
}

func (c *Cluster) checkVersion(ctx context.Context) error {
	rev := c.ListVersion()

	ctx = internal.WithSpec(ctx, specFor("Version", nil))
	if rev.Major() != tolerableMajor || rev.Minor() < tolerableMinor {
		c.AddCode(ctx, 405)
	} else {
		c.AddCode(ctx, 406)
	}

	return nil
}
