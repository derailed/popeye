package internal

import (
	"context"
)

// PopeyeKey namespaces popeye context keys.
type PopeyeKey string

// RunInfo describes a sanitizer run.
type RunInfo struct {
	Section, FQN, Group string
}

const (
	// KeyRun stores run information.
	KeyRun PopeyeKey = "runinfo"
)

// WithGroup adds a group to the context.
func WithGroup(ctx context.Context, grp string) context.Context {
	r := MustExtractRunInfo(ctx)
	r.Group = grp
	return context.WithValue(ctx, KeyRun, r)
}

// WithFQN adds a fqn to the context.
func WithFQN(ctx context.Context, fqn string) context.Context {
	r := MustExtractRunInfo(ctx)
	r.FQN = fqn
	return context.WithValue(ctx, KeyRun, r)
}

// MustExtractFQN extract fqn from context or die.
func MustExtractFQN(ctx context.Context) string {
	r := MustExtractRunInfo(ctx)
	return r.FQN
}

// MustExtractSection extract section from context or die.
func MustExtractSection(ctx context.Context) string {
	r := MustExtractRunInfo(ctx)
	return r.Section
}

// MustExtractRunInfo extracts runinfo from context or die.
func MustExtractRunInfo(ctx context.Context) RunInfo {
	r, ok := ctx.Value(KeyRun).(RunInfo)
	if !ok {
		panic("Doh! No RunInfo in context")
	}
	return r
}
