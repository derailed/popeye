package sanitize

import (
	"testing"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestContainerStatusSanitize(t *testing.T) {
	uu := map[string]struct {
		cs     v1.ContainerStatus
		issues int
		issue  issues.Issue
	}{
		"cool": {
			v1.ContainerStatus{
				Name:         "c1",
				Ready:        true,
				RestartCount: 0,
				State:        v1.ContainerState{},
			},
			0,
			issues.Blank,
		},
		"notReady": {
			v1.ContainerStatus{
				Name:         "c1",
				Ready:        false,
				RestartCount: 0,
				State:        v1.ContainerState{},
			},
			1,
			issues.New(client.NewGVR("containers"), "c1", config.ErrorLevel, "[POP-204] Pod is not ready [0/1]"),
		},
		"waitingNoReason": {
			v1.ContainerStatus{
				Name:         "c1",
				Ready:        false,
				RestartCount: 0,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{Reason: "blah", Message: "blah"},
				},
			},
			1,
			issues.New(client.NewGVR("containers"), "c1", config.ErrorLevel, "[POP-203] Pod is waiting [0/1] blah"),
		},
		"waiting": {
			v1.ContainerStatus{
				Name:         "c1",
				Ready:        false,
				RestartCount: 0,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{},
				},
			},
			1,
			issues.New(client.NewGVR("containers"), "c1", config.ErrorLevel, "[POP-202] Pod is waiting [0/1]"),
		},
		"terminatedReason": {
			v1.ContainerStatus{
				Name:         "c1",
				Ready:        true,
				RestartCount: 0,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{Reason: "blah"},
				},
			},
			1,
			issues.New(client.NewGVR("containers"), "c1", config.WarnLevel, "[POP-201] Pod is terminating [1/1] blah"),
		},
		"terminated": {
			v1.ContainerStatus{
				Name:         "c1",
				Ready:        true,
				RestartCount: 0,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{},
				},
			},
			1,
			issues.New(client.NewGVR("containers"), "c1", config.WarnLevel, "[POP-200] Pod is terminating [1/1]"),
		},
		"terminatedNotReady": {
			v1.ContainerStatus{
				Name:         "c1",
				Ready:        false,
				RestartCount: 0,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{},
				},
			},
			0,
			issues.Blank,
		},
		"restartedLimit": {
			v1.ContainerStatus{
				Name:         "c1",
				Ready:        true,
				RestartCount: 11,
			},
			1,
			issues.New(client.NewGVR("containers"), "c1", config.WarnLevel, "[POP-205] Pod was restarted (11) times"),
		},
	}

	ctx := makeContext("containers", "containers")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := issues.NewCollector(loadCodes(t), makeConfig(t))
			cs := newContainerStatus(c, "default/p1", 1, false, 10)
			cs.sanitize(ctx, u.cs)

			assert.Equal(t, u.issues, len(c.Outcome()["default/p1"]))
			if u.issues != 0 {
				assert.Equal(t, u.issue, c.Outcome()["default/p1"][0])
			}
		})
	}
}
