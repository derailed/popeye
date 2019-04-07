package linter

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func BenchmarkContainerCheckImageTag(b *testing.B) {
	co := v1.Container{
		Name:  "c1",
		Image: "blee",
	}
	l := NewContainer(nil, nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.checkImageTags(co)
	}
}
