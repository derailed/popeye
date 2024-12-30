// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/lint"
	v1 "k8s.io/api/core/v1"
)

// ConfigMap represents a configMap scruber.
type ConfigMap struct {
	*issues.Collector
	*Cache
}

// NewConfigMap returns a new instance.
func NewConfigMap(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &ConfigMap{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *ConfigMap) Preloads() Preloads {
	return Preloads{
		internal.CM: db.LoadResource[*v1.ConfigMap],
		internal.PO: db.LoadResource[*v1.Pod],
	}
}

// Lint all available ConfigMaps.
func (s *ConfigMap) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewConfigMap(s.Collector, s.DB).Lint(ctx)
}
