// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/types"
	v1 "k8s.io/api/core/v1"
)

// ContainerStatus represents container health counts.
type containerStatus struct {
	ready         int
	waiting       int
	terminated    int
	restarts      int
	reason        string
	isInit        bool
	restartsLimit int
	collector     Collector
	fqn           string
	count         int
}

func newContainerStatus(c Collector, fqn string, count int, isInit bool, restarts int) *containerStatus {
	return &containerStatus{
		collector:     c,
		fqn:           fqn,
		isInit:        isInit,
		count:         count,
		restartsLimit: restarts,
	}
}

func (c *containerStatus) sanitize(ctx context.Context, s v1.ContainerStatus) {
	ctx = internal.WithGroup(ctx, types.NewGVR("containers"), s.Name)

	c.rollup(s)
	if c.terminated > 0 && c.ready == 0 {
		return
	}
	if c.terminated > 0 && c.ready != 0 && !c.isInit {
		c.checkReason(ctx, 200, c.reason)
		return
	}
	if c.waiting > 0 {
		c.checkReason(ctx, 202, c.reason)
		return
	}
	if c.ready == 0 {
		c.collector.AddSubCode(ctx, 204, c.ready, c.count)
		return
	}
	if c.restarts > c.restartsLimit {
		c.collector.AddSubCode(ctx, 205, c.restarts, pluralOf("time", c.restarts))
	}
}

func (c *containerStatus) rollup(s v1.ContainerStatus) {
	if s.Ready {
		c.ready++
	}
	if s.State.Waiting != nil {
		c.waiting++
		c.reason = s.State.Waiting.Reason
	}
	if s.State.Terminated != nil {
		c.terminated++
		c.reason = s.State.Terminated.Reason
	}
	c.restarts += int(s.RestartCount)
}

func (c *containerStatus) checkReason(ctx context.Context, code rules.ID, reason string) {
	if reason == "" {
		c.collector.AddSubCode(ctx, code, c.ready, c.count)
		return
	}
	c.collector.AddSubCode(ctx, rules.ID(code+1), c.ready, c.count, c.reason)
}
