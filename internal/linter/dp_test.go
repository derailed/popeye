package linter

import (
	"context"
	"fmt"
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

func TestDPLinter(t *testing.T) {
	mkl := NewMockLoader()
	m.When(mkl.ListDeployments()).ThenReturn(map[string]appsv1.Deployment{
		"default/dp1": makeDP("dp1", "100m", "10Mi"),
		"default/dp2": makeDP("dp2", "100m", "10Mi"),
	}, nil)

	dp := NewDeployment(mkl, nil)
	dp.Lint(context.Background())

	assert.Equal(t, 2, len(dp.Issues()))
	mkl.VerifyWasCalledOnce().ListDeployments()
}

func TestDPLint(t *testing.T) {
	uu := []struct {
		dps    map[string]appsv1.Deployment
		issues int
	}{
		{map[string]appsv1.Deployment{
			"default/dp1": makeDP("dp1", "100m", "10Mi"),
			"default/dp2": makeDP("dp2", "100m", "10Mi"),
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
		dp := NewDeployment(mkl, nil)
		dp.lint(u.dps)

		assert.Equal(t, 2, len(dp.Issues()))
		assert.Equal(t, 2, len(dp.Issues()["default/dp1"][0].SubIssues()))
		assert.Equal(t, 2, len(dp.Issues()["default/dp2"][0].SubIssues()))
		mkl.VerifyWasCalled(pegomock.Times(2)).ListPodsByLabels("fred=blee")
	}
}

func TestDPCheckContainers(t *testing.T) {
	uu := []struct {
		dp     appsv1.Deployment
		level  Level
		issues int
	}{
		{makeDP("dp1", "100m", "5Mi"), InfoLevel, 2},
	}

	fqn := "default/dp1"
	for _, u := range uu {
		dp := NewDeployment(nil, nil)
		dp.checkContainers(metaFQN(u.dp.ObjectMeta), u.dp)

		assert.Equal(t, u.issues, len(dp.Issues()[fqn][0].SubIssues()))
		if u.issues == 1 {
			assert.Equal(t, u.level, dp.Issues()[fqn][0].Severity())
		}
	}
}

func TestDPCheckDeployment(t *testing.T) {
	dp1 := makeDP("dp1", "100m", "5Mi")
	dp1.Spec.Replicas = new(int32)

	dp2 := makeDP("dp2", "100m", "5Mi")
	dp2.Status.AvailableReplicas = 0

	dp3 := makeDP("dp3", "100m", "5Mi")
	dp3.Status.CollisionCount = new(int32)

	uu := []struct {
		key    string
		dp     appsv1.Deployment
		level  Level
		issues int
	}{
		{"default/dp", makeDP("dp", "100m", "1Mi"), InfoLevel, 0},
		{"default/dp1", dp1, InfoLevel, 1},
		{"default/dp2", dp2, WarnLevel, 1},
		{"default/dp3", dp3, ErrorLevel, 1},
	}

	for _, u := range uu {
		fqn := fqn("default", u.key)
		dp := NewDeployment(nil, nil)
		dp.checkDeployment(fqn, u.dp)

		assert.Equal(t, u.issues, len(dp.Issues()[fqn]))
		if u.issues == 1 {
			assert.Equal(t, u.level, dp.Issues()[fqn][0].Severity())
		}
	}
}

func TestDPCheckUtilization(t *testing.T) {
	uu := []struct {
		dp       appsv1.Deployment
		cpu, mem string
		issues   int
	}{
		// All Good!
		{makeDP("dp1", "100m", "10Mi"), "200m", "20Mi", 0},
		// Over CPU
		{makeDP("dp1", "1", "10Mi"), "100m", "20Mi", 1},
		// Under CPU
		{makeDP("dp1", "15m", "10Mi"), "100m", "20Mi", 1},
		// Over MEM
		{makeDP("dp1", "100m", "10Mi"), "200m", "10Mi", 1},
		// Under MEM
		{makeDP("dp1", "100m", "2Mi"), "200m", "10Mi", 1},
	}

	mkl := NewMockLoader()
	m.When(mkl.ListPodsByLabels("fred=blee")).ThenReturn(map[string]v1.Pod{
		"default/p1": makePod("p1"),
	}, nil)
	m.When(mkl.ClusterHasMetrics()).ThenReturn(true, nil)
	m.When(mkl.CPUResourceLimits()).ThenReturn(config.Allocations{100, 50})
	m.When(mkl.MEMResourceLimits()).ThenReturn(config.Allocations{100, 50})

	fqn := fqn("default", "dp1")
	for _, u := range uu {
		pmx := k8s.PodsMetrics{
			"default/p1": makeContainerMx("c1", u.cpu, u.mem),
		}

		dp := NewDeployment(mkl, nil)
		dp.checkUtilization(fqn, u.dp, pmx)

		assert.Equal(t, u.issues, len(dp.Issues()[fqn]))
	}
	mkl.VerifyWasCalled(pegomock.Times(5)).ListPodsByLabels("fred=blee")
}

// ----------------------------------------------------------------------------
// Helpers...

func makeContainerMx(n, cpu, mem string) k8s.ContainerMetrics {
	return k8s.ContainerMetrics{
		n: k8s.Metrics{
			CurrentCPU: toQty(cpu),
			CurrentMEM: toQty(mem),
		},
	}
}

func makeDP(n, cpu, mem string) appsv1.Deployment {
	var count int32 = 1

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
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
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 1,
		},
	}
}

func dumpIssues(ii Issues) {
	for k, v := range ii {
		fmt.Println(k)
		for _, i := range v {
			if i.HasSubIssues() {
				for kk, vv := range i.SubIssues() {
					fmt.Printf("  %s  %v\n", kk, vv)
				}
			} else {
				fmt.Printf("  %v\n", i)
			}
		}
	}
}
