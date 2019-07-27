package sanitize

import (
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

func (c *containerStatus) sanitize(s v1.ContainerStatus) {
	c.rollup(s)

	if c.terminated > 0 && c.ready == 0 {
		return
	}

	if c.terminated > 0 && c.ready != 0 && !c.isInit {
		if c.reason == "" {
			c.collector.AddSubCode(200, c.fqn, s.Name, c.ready, c.count)
		} else {
			c.collector.AddSubCode(201, c.fqn, s.Name, c.ready, c.count, c.reason)
		}
		return
	}

	if c.waiting > 0 {
		if c.reason == "" {
			c.collector.AddSubCode(202, c.fqn, s.Name, c.ready, c.count)
		} else {
			c.collector.AddSubCode(203, c.fqn, s.Name, c.ready, c.count, c.reason)
		}
		return
	}

	if c.ready == 0 {
		c.collector.AddSubCode(204, c.fqn, s.Name, c.ready, c.count)
		return
	}

	if c.restarts > c.restartsLimit {
		c.collector.AddSubCode(205, c.fqn, s.Name, c.restarts, pluralOf("time", c.restarts))
	}

	return
}
