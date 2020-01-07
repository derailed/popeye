package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
)

// Sanitizer represents a resource sanitizer.
type Sanitizer interface {
	Collector
	Sanitize(context.Context) error
}

// Collector collects sanitization issues.
type Collector interface {
	MaxSeverity(res string) config.Level
	Outcome() issues.Outcome
}
