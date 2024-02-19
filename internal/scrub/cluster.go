// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/lint"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog/log"
)

// Cluster represents a Cluster scruber.
type Cluster struct {
	*issues.Collector
	*cache.Cluster
	*config.Config

	client types.Connection
}

// NewCluster returns a new instance.
func NewCluster(ctx context.Context, c *Cache, codes *issues.Codes) Linter {
	cl := Cluster{
		client:    c.factory.Client(),
		Config:    c.Config,
		Collector: issues.NewCollector(codes, c.Config),
	}

	var err error
	cl.Cluster, err = c.cluster(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to gather cluster info")
	}

	return &cl
}

func (d *Cluster) Preloads() Preloads {
	return nil
}

// Lint all available Clusters.
func (d *Cluster) Lint(ctx context.Context) error {
	return lint.NewCluster(d.Collector, d).Lint(ctx)
}

func (d *Cluster) HasMetrics() bool {
	return d.client.HasMetrics()
}
