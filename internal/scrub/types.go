// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/types"
)

type Scrubs map[internal.R]ScrubFn

// ScrubFn represents a resource scruber.
type ScrubFn func(context.Context, *Cache, *issues.Codes) Linter

// LoaderFn represents a resource loader.
type LoaderFn func(context.Context, *db.Loader, types.GVR) error

// Collector collects sanitization issues.
type Collector interface {
	MaxSeverity(res string) rules.Level
	Outcome() issues.Outcome
}

// Linter represents a resource linter.
type Linter interface {
	// Collector tracks issues.
	Collector

	// Lint runs checks on a resource.
	Lint(context.Context) error

	// Preloads Preloads resource requirements.
	Preloads() Preloads
}
