package linter

import (
	"context"
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestHPALinter(t *testing.T) {
	mkl := NewMockLoader()
	m.When(mkl.ListHorizontalPodAutoscalers()).ThenReturn(map[string]autoscalingv1.HorizontalPodAutoscaler{
		"default/h1": makeHPA("h1", "Deployment", "d1", 1),
	}, nil)
	m.When(mkl.ListDeployments()).ThenReturn(map[string]appsv1.Deployment{
		"default/d1": makeDP("d1", "100m", "10Mi"),
	}, nil)
	m.When(mkl.ClusterHasMetrics()).ThenReturn(true, nil)
	m.When(mkl.ListNodes()).ThenReturn([]v1.Node{
		makeNodeMX("n1", "8", "8Gi", "5", "5Gi"),
	}, nil)
	m.When(mkl.FetchNodesMetrics()).ThenReturn([]mv1beta1.NodeMetrics{
		makeMxNode("n1", "8", "8Gi"),
	}, nil)

	hpa := NewHorizontalPodAutoscaler(mkl, nil)
	hpa.Lint(context.Background())

	assert.Equal(t, 2, len(hpa.Issues()))
	mkl.VerifyWasCalledOnce().ListHorizontalPodAutoscalers()
	mkl.VerifyWasCalledOnce().ListDeployments()
	mkl.VerifyWasCalledOnce().ListNodes()
	mkl.VerifyWasCalledOnce().FetchNodesMetrics()
}

func TestHPALintDP(t *testing.T) {
	uu := []struct {
		hpas   map[string]autoscalingv1.HorizontalPodAutoscaler
		dps    map[string]appsv1.Deployment
		keys   map[string]int
		issues int
	}{
		// Happy!
		{
			map[string]autoscalingv1.HorizontalPodAutoscaler{
				"default/h1": makeHPA("h1", "Deployment", "d1", 1),
			},
			map[string]appsv1.Deployment{
				"default/d1": makeDP("d1", "100m", "10Mi"),
			},
			map[string]int{
				"default/h1": 0,
			},
			0,
		},
		// Over 1c + 20Mi
		{
			map[string]autoscalingv1.HorizontalPodAutoscaler{
				"default/h1": makeHPA("h1", "Deployment", "d1", 2),
			},
			map[string]appsv1.Deployment{
				"default/d1": makeDP("d1", "1000m", "20Mi"),
			},
			map[string]int{
				"default/h1": 2,
			},
			2,
		},
		// One cool, one not -- Over 30Mi
		{
			map[string]autoscalingv1.HorizontalPodAutoscaler{
				"default/h1": makeHPA("h1", "Deployment", "d1", 1),
				"default/h2": makeHPA("h2", "Deployment", "d2", 5),
			},
			map[string]appsv1.Deployment{
				"default/d1": makeDP("d1", "1m", "1Mi"),
				"default/d2": makeDP("d2", "100m", "10Mi"),
			},
			map[string]int{
				"default/h1": 0,
				"default/h2": 1,
			},
			1,
		},
		// One cool, one not -- Over 1c
		{
			map[string]autoscalingv1.HorizontalPodAutoscaler{
				"default/h1": makeHPA("h1", "Deployment", "d1", 1),
				"default/h2": makeHPA("h2", "Deployment", "d2", 5),
			},
			map[string]appsv1.Deployment{
				"default/d1": makeDP("d1", "1m", "1Mi"),
				"default/d2": makeDP("d2", "400m", "1Mi"),
			},
			map[string]int{
				"default/h1": 0,
				"default/h2": 1,
			},
			1,
		},
	}

	for _, u := range uu {
		h := NewHorizontalPodAutoscaler(nil, nil)
		var sts map[string]appsv1.StatefulSet
		h.lint(u.hpas, u.dps, sts, toQty("1"), toQty("20Mi"))

		for k, v := range u.keys {
			assert.Equal(t, v, len(h.Issues()[k]))
		}
		assert.Equal(t, u.issues, len(h.Issues()["hpas"]))
	}
}

func TestHPALintSTS(t *testing.T) {
	uu := []struct {
		hpas   map[string]autoscalingv1.HorizontalPodAutoscaler
		sts    map[string]appsv1.StatefulSet
		keys   map[string]int
		issues int
	}{
		// Happy!
		{
			map[string]autoscalingv1.HorizontalPodAutoscaler{
				"default/h1": makeHPA("h1", "StatefulSet", "s1", 1),
			},
			map[string]appsv1.StatefulSet{
				"default/s1": makeSTS("s1", "100m", "10Mi"),
			},
			map[string]int{
				"default/h1": 0,
			},
			0,
		},
		// Over 1c + 20Mi
		{
			map[string]autoscalingv1.HorizontalPodAutoscaler{
				"default/h1": makeHPA("h1", "StatefulSet", "s1", 2),
			},
			map[string]appsv1.StatefulSet{
				"default/s1": makeSTS("s1", "1000m", "20Mi"),
			},
			map[string]int{
				"default/h1": 2,
			},
			2,
		},
		// One cool, one not -- Over 30Mi
		{
			map[string]autoscalingv1.HorizontalPodAutoscaler{
				"default/h1": makeHPA("h1", "StatefulSet", "s1", 1),
				"default/h2": makeHPA("h2", "StatefulSet", "s2", 5),
			},
			map[string]appsv1.StatefulSet{
				"default/s1": makeSTS("s1", "1m", "1Mi"),
				"default/s2": makeSTS("s2", "100m", "10Mi"),
			},
			map[string]int{
				"default/h1": 0,
				"default/h2": 1,
			},
			1,
		},
		// One cool, one not -- Over 1c
		{
			map[string]autoscalingv1.HorizontalPodAutoscaler{
				"default/h1": makeHPA("h1", "StatefulSet", "s1", 1),
				"default/h2": makeHPA("h2", "StatefulSet", "s2", 5),
			},
			map[string]appsv1.StatefulSet{
				"default/s1": makeSTS("s1", "1m", "1Mi"),
				"default/s2": makeSTS("s2", "400m", "1Mi"),
			},
			map[string]int{
				"default/h1": 0,
				"default/h2": 1,
			},
			1,
		},
	}

	for _, u := range uu {
		h := NewHorizontalPodAutoscaler(nil, nil)
		var dps map[string]appsv1.Deployment
		h.lint(u.hpas, dps, u.sts, toQty("1"), toQty("20Mi"))

		for k, v := range u.keys {
			assert.Equal(t, v, len(h.Issues()[k]))
		}
		assert.Equal(t, u.issues, len(h.Issues()["hpas"]))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeHPA(n, kind, dp string, max int32) autoscalingv1.HorizontalPodAutoscaler {
	return autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			MaxReplicas: max,
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind: kind,
				Name: dp,
			},
		},
	}
}
