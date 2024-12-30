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

// PersistentVolume represents a PersistentVolume scruber.
type PersistentVolume struct {
	*issues.Collector
	*Cache
}

// NewPersistentVolume return a new instance.
func NewPersistentVolume(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &PersistentVolume{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *PersistentVolume) Preloads() Preloads {
	return Preloads{
		internal.PV: db.LoadResource[*v1.PersistentVolume],
		internal.PO: db.LoadResource[*v1.Pod],
	}
}

// Lint all available PersistentVolumes.
func (s *PersistentVolume) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewPersistentVolume(s.Collector, s.DB).Lint(ctx)
}
