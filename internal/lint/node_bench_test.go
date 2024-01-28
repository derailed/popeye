// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

// import (
// 	"testing"

// 	"github.com/derailed/popeye/internal/client"
// 	v1 "k8s.io/api/core/v1"
// )

// func BenchmarkNodeTaints(b *testing.B) {
// 	no := makeTaintedNode("n1")
// 	tt := tolerations{
// 		"duh:f1":  struct{}{},
// 		"blee:f2": struct{}{},
// 	}

// 	l := NewNode(nil, nil)
// 	b.ResetTimer()
// 	b.ReportAllocs()
// 	for n := 0; n < b.N; n++ {
// 		l.checkTaints(no, tt)
// 	}
// }

// func BenchmarkNodeLint(b *testing.B) {
// 	no := makeCondNode("n1", v1.NodeReady, v1.ConditionFalse)
// 	tt := tolerations{
// 		"duh:f1":  struct{}{},
// 		"blee:f2": struct{}{},
// 	}

// 	l := NewNode(nil, nil)
// 	b.ResetTimer()
// 	b.ReportAllocs()
// 	for n := 0; n < b.N; n++ {
// 		l.lint(no, k8s.NodeMetrics{}, tt)
// 	}
// }
