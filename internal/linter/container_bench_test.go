package linter

import (
	"fmt"
	"testing"

	m "github.com/petergtz/pegomock"
	v1 "k8s.io/api/core/v1"
)

func TestSetup(t *testing.T) {
	m.RegisterMockTestingT(t)
	m.RegisterMockFailHandler(func(m string, i ...int) {
		fmt.Println("Boom!", m, i)
	})
}

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
