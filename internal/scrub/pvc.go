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

// PersistentVolumeClaim represents a PersistentVolumeClaim scruber.
type PersistentVolumeClaim struct {
	*issues.Collector
	*Cache
}

// NewPersistentVolumeClaim returns a new instance.
func NewPersistentVolumeClaim(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &PersistentVolumeClaim{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *PersistentVolumeClaim) Preloads() Preloads {
	return Preloads{
		internal.PVC: db.LoadResource[*v1.PersistentVolumeClaim],
		internal.PO:  db.LoadResource[*v1.Pod],
	}
}

// Lint all available PersistentVolumeClaims.
func (s *PersistentVolumeClaim) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewPersistentVolumeClaim(s.Collector, s.DB).Lint(ctx)
}
