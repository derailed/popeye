package linter

import (
	v1 "k8s.io/api/core/v1"
)

// ContainerStatusCounts represents container health counts.
type containerStatusCount struct {
	ready      int
	waiting    int
	terminated int
	restarts   int
}

func (c *containerStatusCount) rollup(s v1.ContainerStatus) {
	if s.Ready {
		c.ready++
	}

	if s.State.Waiting != nil {
		c.waiting++
	}

	if s.State.Terminated != nil {
		c.terminated++
	}

	c.restarts += int(s.RestartCount)
}

func (c *containerStatusCount) diagnose(total int) Issue {
	if c.terminated > 0 {
		return NewErrorf(WarnLevel, "Pod is terminating (%d/%d)", c.terminated, total)
	}

	if c.waiting > 0 {
		return NewErrorf(WarnLevel, "Pod is waiting (%d/%d)", c.waiting, total)
	}

	if c.ready == 0 {
		return NewErrorf(ErrorLevel, "Pod is not ready (%d/%d)", c.ready, total)
	}

	if c.restarts > 0 {
		return NewErrorf(WarnLevel, "Pod was restarted (%d) %s", c.restarts, pluralOf("time", c.restarts))
	}

	return nil
}

// Poor man plural...
func pluralOf(s string, count int) string {
	if count > 1 {
		return s + "s"
	}
	return s
}
