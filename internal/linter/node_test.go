package linter

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	m "github.com/petergtz/pegomock"
	pegomock "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNoLinter(t *testing.T) {
	mks := NewMockClient()
	m.When(mks.ListNodes()).ThenReturn([]v1.Node{
		makeCondNode("n1", v1.NodeReady, v1.ConditionFalse),
		makeNode("n2"),
	}, nil)
	m.When(mks.ClusterHasMetrics()).ThenReturn(true)
	m.When(mks.FetchNodesMetrics()).ThenReturn([]mv1beta1.NodeMetrics{
		makeMxNode("n1", "50m", "1Mi"),
		makeMxNode("n2", "50m", "1Mi"),
	}, nil)

	l := NewNode(mks, nil)
	l.Lint(context.Background())

	assert.Equal(t, 2, len(l.Issues()))
	assert.Equal(t, 1, len(l.Issues()["n1"]))
	assert.Equal(t, 0, len(l.Issues()["n2"]))

	mks.VerifyWasCalled(pegomock.Times(1)).ListNodes()
}

func TestNodeLint(t *testing.T) {
	uu := []struct {
		no     v1.Node
		issues int
	}{
		{
			no:     makeCondNode("n1", v1.NodeReady, v1.ConditionFalse),
			issues: 1,
		},
		{
			no:     makeCondNode("n1", v1.NodeReady, v1.ConditionTrue),
			issues: 1,
		},
	}

	for _, u := range uu {
		l := NewNode(nil, nil)
		l.lint(u.no, k8s.NodeMetrics{}, tolerations{})
		assert.Equal(t, u.issues, len(l.Issues()[u.no.Name]))
	}
}

func TestNodeUtilization(t *testing.T) {
	uu := []struct {
		mx     k8s.NodeMetrics
		issues int
		level  Level
	}{
		{
			mx:     k8s.NodeMetrics{CurrentCPU: 500, AvailCPU: 1000, CurrentMEM: 1000, AvailMEM: 2000},
			issues: 0,
		},
		{
			mx:     k8s.NodeMetrics{CurrentCPU: 900, AvailCPU: 1000, CurrentMEM: 1000, AvailMEM: 2000},
			issues: 1,
			level:  WarnLevel,
		},
		{
			mx:     k8s.NodeMetrics{CurrentCPU: 500, AvailCPU: 1000, CurrentMEM: 9000, AvailMEM: 10000},
			issues: 1,
			level:  WarnLevel,
		},
		{
			mx:     k8s.NodeMetrics{CurrentCPU: 900, AvailCPU: 1000, CurrentMEM: 9000, AvailMEM: 10000},
			issues: 2,
			level:  WarnLevel,
		},
		{
			mx:     k8s.NodeMetrics{},
			issues: 1,
			level:  WarnLevel,
		},
	}

	for _, u := range uu {
		l := NewNode(k8s.NewClient(config.New()), nil)
		l.checkUtilization("n1", u.mx)
		assert.Equal(t, u.issues, len(l.Issues()["n1"]))
	}
}

func TestNodeCheckConditions(t *testing.T) {
	uu := []struct {
		no     v1.Node
		issues int
		ready  bool
	}{
		{
			no:     makeCondNode("n1", v1.NodeReady, v1.ConditionTrue),
			issues: 0,
			ready:  true,
		},
		{
			no:     makeCondNode("n1", v1.NodeOutOfDisk, v1.ConditionTrue),
			issues: 1,
		},
		{
			no:     makeCondNode("n1", v1.NodeOutOfDisk, v1.ConditionFalse),
			issues: 0,
		},
		{
			no:     makeCondNode("n1", v1.NodeReady, v1.ConditionFalse),
			issues: 1,
		},
		{
			no:     makeCondNode("n1", v1.NodeReady, v1.ConditionFalse),
			issues: 1,
		},
		{
			no:     makeCondNode("n1", v1.NodeReady, v1.ConditionUnknown),
			issues: 1,
		},
		{
			no:     makeHosedNode("n1"),
			issues: 6,
		},
	}

	for _, u := range uu {
		l := NewNode(nil, nil)
		assert.Equal(t, u.ready, l.checkConditions(u.no))
		assert.Equal(t, u.issues, len(l.Issues()[u.no.Name]))
	}
}

func TestNodeCheckTaints(t *testing.T) {
	uu := []struct {
		no     v1.Node
		tt     tolerations
		issues int
	}{
		{
			no: makeTaintedNode("n1"),
			tt: tolerations{
				"fred:f1": struct{}{},
				"blee:f2": struct{}{},
			},
			issues: 0,
		},
		{
			no: makeTaintedNode("n1"),
			tt: tolerations{
				"duh:f1":  struct{}{},
				"blee:f2": struct{}{},
			},
			issues: 1,
		},
	}

	mks := NewMockClient()
	m.When(mks.ListPods()).ThenReturn(map[string]v1.Pod{
		"default/p1": makePod("p1"),
		"default/p2": makePod("p2"),
	}, nil)

	for _, u := range uu {
		l := NewNode(mks, nil)
		l.checkTaints(u.no, u.tt)
		assert.Equal(t, u.issues, len(l.Issues()[u.no.Name]))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeCondNode(n string, c v1.NodeConditionType, s v1.ConditionStatus) v1.Node {
	no := makeNode(n)
	no.Status.Conditions = append(no.Status.Conditions,
		v1.NodeCondition{Type: c, Status: s},
	)
	return no
}

func makeHosedNode(n string) v1.Node {
	no := makeNode(n)
	no.Status.Conditions = append(no.Status.Conditions,
		v1.NodeCondition{
			Type:   v1.NodeOutOfDisk,
			Status: v1.ConditionTrue,
		},
		v1.NodeCondition{
			Type:   v1.NodeMemoryPressure,
			Status: v1.ConditionTrue,
		},
		v1.NodeCondition{
			Type:   v1.NodeDiskPressure,
			Status: v1.ConditionTrue,
		},
		v1.NodeCondition{
			Type:   v1.NodeMemoryPressure,
			Status: v1.ConditionTrue,
		},
		v1.NodeCondition{
			Type:   v1.NodePIDPressure,
			Status: v1.ConditionTrue,
		},
		v1.NodeCondition{
			Type:   v1.NodeNetworkUnavailable,
			Status: v1.ConditionTrue,
		},
	)
	return no
}

func makeNode(n string) v1.Node {
	return v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
		Spec:   v1.NodeSpec{},
		Status: v1.NodeStatus{},
	}
}

func makeTaintedNode(n string) v1.Node {
	no := makeNode(n)
	no.Spec.Taints = []v1.Taint{
		{Key: "fred", Value: "f1"},
		{Key: "blee", Value: "f2"},
	}
	return no
}

func makeMxNode(name, cpu, mem string) v1beta1.NodeMetrics {
	return v1beta1.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Usage: makeRes(cpu, mem),
	}
}
