package linter

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	m "github.com/petergtz/pegomock"
	pegomock "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestSTSLinter(t *testing.T) {
	mkl := NewMockLoader()
	m.When(mkl.ListStatefulSets()).ThenReturn(map[string]appsv1.StatefulSet{
		"default/st1": makeSTS("st1", "100m", "10Mi"),
		"default/st2": makeSTS("st2", "100m", "10Mi"),
	}, nil)

	sts := NewStatefulSet(mkl, nil)
	sts.Lint(context.Background())

	assert.Equal(t, 2, len(sts.Issues()))
	mkl.VerifyWasCalledOnce().ListStatefulSets()
}

func TestSTSLint(t *testing.T) {
	uu := []struct {
		stss   map[string]appsv1.StatefulSet
		issues int
	}{
		{
			map[string]appsv1.StatefulSet{
				"default/sts1": makeSTS("sts1", "100m", "10Mi"),
				"default/sts2": makeSTS("sts2", "100m", "10Mi"),
			},
			2,
		},
	}

	mkl := NewMockLoader()
	m.When(mkl.ListPodsByLabels("fred=blee")).ThenReturn(map[string]v1.Pod{
		"default/p1": makePod("p1"),
		"default/p2": makePod("p2"),
	}, nil)

	m.When(mkl.ClusterHasMetrics()).ThenReturn(true, nil)
	m.When(mkl.FetchPodsMetrics("")).ThenReturn([]mv1beta1.PodMetrics{
		makeMxPod("p1", "100m", "10Mi"),
		makeMxPod("p2", "100m", "10Mi"),
	}, nil)

	for _, u := range uu {
		sts := NewStatefulSet(mkl, nil)
		sts.lint(u.stss)

		assert.Equal(t, 2, len(sts.Issues()))
		assert.Equal(t, 2, len(sts.Issues()["default/sts1"][0].SubIssues()))
		assert.Equal(t, 2, len(sts.Issues()["default/sts2"][0].SubIssues()))
		mkl.VerifyWasCalled(pegomock.Times(2)).ListPodsByLabels("fred=blee")
	}
}

func TestSTSCheckContainers(t *testing.T) {
	uu := []struct {
		sts    appsv1.StatefulSet
		level  Level
		issues int
	}{
		{makeSTS("sts1", "100m", "5Mi"), InfoLevel, 2},
	}

	fqn := "default/sts1"
	for _, u := range uu {
		sts := NewStatefulSet(nil, nil)
		sts.checkContainers(metaFQN(u.sts.ObjectMeta), u.sts)

		assert.Equal(t, u.issues, len(sts.Issues()[fqn][0].SubIssues()))
		if u.issues == 1 {
			assert.Equal(t, u.level, sts.Issues()[fqn][0].Severity())
		}
	}
}

func TestSTSCheckStatefulSet(t *testing.T) {
	sts1 := makeSTS("sts1", "100m", "5Mi")
	sts1.Spec.Replicas = new(int32)

	sts2 := makeSTS("sts2", "100m", "5Mi")
	sts2.Status.CurrentReplicas = 0

	sts3 := makeSTS("sts3", "100m", "5Mi")
	var count int32 = 1
	sts3.Status.CollisionCount = &count

	uu := []struct {
		key    string
		sts    appsv1.StatefulSet
		level  Level
		issues int
	}{
		{"default/sts", makeSTS("sts", "100m", "1Mi"), InfoLevel, 0},
		{"default/sts1", sts1, InfoLevel, 1},
		{"default/sts2", sts2, WarnLevel, 1},
		{"default/sts3", sts3, ErrorLevel, 1},
	}

	for _, u := range uu {
		fqn := fqn("default", u.key)
		sts := NewStatefulSet(nil, nil)
		sts.checkStatefulSet(fqn, u.sts)

		assert.Equal(t, u.issues, len(sts.Issues()))
		if len(sts.Issues()) > 0 {
			assert.Equal(t, u.level, sts.Issues()[fqn][0].Severity())
		}
	}
}

func TestSTSCheckUtilization(t *testing.T) {
	uu := []struct {
		sts      appsv1.StatefulSet
		cpu, mem string
		issues   int
	}{
		// All Good!
		{makeSTS("sts1", "100m", "10Mi"), "200m", "20Mi", 0},
		// Over CPU
		{makeSTS("sts1", "1", "10Mi"), "100m", "20Mi", 1},
		// Under CPU
		{makeSTS("sts1", "15m", "10Mi"), "100m", "20Mi", 1},
		// Over MEM
		{makeSTS("sts1", "100m", "10Mi"), "200m", "10Mi", 1},
		// Under MEM
		{makeSTS("sts1", "100m", "2Mi"), "200m", "10Mi", 1},
	}

	mkl := NewMockLoader()
	m.When(mkl.ListPodsByLabels("fred=blee")).ThenReturn(map[string]v1.Pod{
		"default/p1": makePod("p1"),
	}, nil)
	m.When(mkl.ClusterHasMetrics()).ThenReturn(true, nil)
	m.When(mkl.CPUResourceLimits()).ThenReturn(config.Allocations{100, 50})
	m.When(mkl.MEMResourceLimits()).ThenReturn(config.Allocations{100, 50})

	fqn := fqn("default", "sts1")
	for _, u := range uu {
		pmx := k8s.PodsMetrics{
			"default/p1": makeContainerMx("c1", u.cpu, u.mem),
		}

		sts := NewStatefulSet(mkl, nil)
		sts.checkUtilization(fqn, u.sts, pmx)

		assert.Equal(t, u.issues, len(sts.Issues()[fqn]))
	}
	mkl.VerifyWasCalled(pegomock.Times(5)).ListPodsByLabels("fred=blee")
}

// ----------------------------------------------------------------------------
// Helpers...

func makeSTS(n, cpu, mem string) appsv1.StatefulSet {
	var count int32 = 1

	return appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &count,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"fred": "blee",
				},
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "c1",
							Image: "fred:0.0.1",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    toQty(cpu),
									v1.ResourceMemory: toQty(mem),
								},
							},
						},
					},
					InitContainers: []v1.Container{
						{
							Name:  "i1",
							Image: "fred:0.0.1",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    toQty(cpu),
									v1.ResourceMemory: toQty(mem),
								},
							},
						},
					},
				},
			},
		},
		Status: appsv1.StatefulSetStatus{
			CurrentReplicas: 1,
		},
	}
}
