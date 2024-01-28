// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func BenchmarkContainerCheckImageTag(b *testing.B) {
	co := v1.Container{
		Name:  "c1",
		Image: "blee",
	}
	l := NewContainer("", nil)

	b.ResetTimer()
	b.ReportAllocs()
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		l.checkImageTags(ctx, co.Image)
	}
}
