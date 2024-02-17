// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package internal

import (
	"context"

	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/types"
)

// RunInfo describes a scan run.
type RunInfo struct {
	Section    string
	SectionGVR types.GVR
	Group      string
	GroupGVR   types.GVR
	Spec       rules.Spec
	Total      int
}

func NewRunInfo(gvr types.GVR) RunInfo {
	return RunInfo{
		Section:    gvr.R(),
		SectionGVR: gvr,
	}
}

// WithGroup adds a group to the context.
func WithGroup(ctx context.Context, gvr types.GVR, grp string) context.Context {
	r := MustExtractRunInfo(ctx)
	r.Group, r.GroupGVR = grp, gvr

	return context.WithValue(ctx, KeyRunInfo, r)
}

func WithSpec(ctx context.Context, spec rules.Spec) context.Context {
	r := MustExtractRunInfo(ctx)
	r.Spec = spec

	return context.WithValue(ctx, KeyRunInfo, r)
}

// MustExtractSectionGVR extract section gvr from context or die.
func MustExtractSectionGVR(ctx context.Context) types.GVR {
	r := MustExtractRunInfo(ctx)
	return r.SectionGVR
}

// MustExtractRunInfo extracts runinfo from context or die.
func MustExtractRunInfo(ctx context.Context) RunInfo {
	r, ok := ctx.Value(KeyRunInfo).(RunInfo)
	if !ok {
		panic("Doh! No RunInfo in context")
	}
	return r
}

func MustExtractFactory(ctx context.Context) types.Factory {
	f, ok := ctx.Value(KeyFactory).(types.Factory)
	if !ok {
		panic("Doh! No factory in context")
	}
	return f
}
