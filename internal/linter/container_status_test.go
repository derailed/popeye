package linter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestContainerStatusRollup(t *testing.T) {
	uu := []struct {
		ready, waiting, terminated bool
		restarts                   int32
		expected                   containerStatusCount
	}{
		{ready: true, restarts: 10, expected: containerStatusCount{1, 0, 0, 10}},
		{waiting: true, restarts: 10, expected: containerStatusCount{0, 1, 0, 10}},
		{terminated: true, restarts: 10, expected: containerStatusCount{0, 0, 1, 10}},
	}

	for _, u := range uu {
		cs := v1.ContainerStatus{
			Ready:        u.ready,
			RestartCount: u.restarts,
		}
		if u.waiting {
			cs.State = v1.ContainerState{Waiting: &v1.ContainerStateWaiting{}}
		}
		if u.terminated {
			cs.State = v1.ContainerState{Terminated: &v1.ContainerStateTerminated{}}
		}

		counts := new(containerStatusCount)
		counts.rollup(cs)

		assert.Equal(t, u.expected.ready, counts.ready)
		assert.Equal(t, u.expected.terminated, counts.terminated)
		assert.Equal(t, u.expected.waiting, counts.waiting)
		assert.Equal(t, u.expected.restarts, counts.restarts)
	}
}

func TestContainerStatusDiagnose(t *testing.T) {
	uu := []struct {
		counts containerStatusCount
		total  int
		issue  Issue
	}{
		{containerStatusCount{0, 0, 0, 0}, 1, NewError(ErrorLevel, "Pod is not ready (0/1)")},
		{containerStatusCount{1, 0, 0, 0}, 1, nil},
		{containerStatusCount{0, 1, 0, 0}, 1, NewError(WarnLevel, "Pod is waiting (1/1)")},
		{containerStatusCount{0, 0, 1, 0}, 1, NewError(WarnLevel, "Pod is terminating (1/1)")},
		{containerStatusCount{1, 0, 0, 1}, 1, NewError(WarnLevel, "Pod was restarted (1) time")},
		{containerStatusCount{1, 0, 0, 10}, 1, NewError(WarnLevel, "Pod was restarted (10) times")},
		{containerStatusCount{1, 0, 0, 0}, 0, nil},
	}

	for _, u := range uu {
		assert.Equal(t, u.issue, u.counts.diagnose(u.total, 0, false))
	}
}
