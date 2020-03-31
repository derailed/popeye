package internal

import (
	"context"

	"github.com/derailed/popeye/internal/client"
)

// RunInfo describes a sanitizer run.
type RunInfo struct {
	Section    string
	SectionGVR client.GVR
	FQN        string
	Group      string
	GroupGVR   client.GVR
}

// WithGroup adds a group to the context.
func WithGroup(ctx context.Context, gvr client.GVR, grp string) context.Context {
	r := MustExtractRunInfo(ctx)
	r.Group, r.GroupGVR = grp, gvr
	return context.WithValue(ctx, KeyRunInfo, r)
}

// WithFQN adds a fqn to the context.
func WithFQN(ctx context.Context, fqn string) context.Context {
	r := MustExtractRunInfo(ctx)
	r.FQN = fqn
	return context.WithValue(ctx, KeyRunInfo, r)
}

// MustExtractFQN extract fqn from context or die.
func MustExtractFQN(ctx context.Context) string {
	r := MustExtractRunInfo(ctx)
	return r.FQN
}

// MustExtractSectionGVR extract section gvr from context or die.
func MustExtractSectionGVR(ctx context.Context) string {
	r := MustExtractRunInfo(ctx)
	return r.SectionGVR.String()
}

// MustExtractRunInfo extracts runinfo from context or die.
func MustExtractRunInfo(ctx context.Context) RunInfo {
	r, ok := ctx.Value(KeyRunInfo).(RunInfo)
	if !ok {
		panic("Doh! No RunInfo in context")
	}
	return r
}
