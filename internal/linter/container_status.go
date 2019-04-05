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
	reason     string
}

func (c *containerStatusCount) rollup(s v1.ContainerStatus) {
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

func (c *containerStatusCount) diagnose(total int, restartsLimit int, isInit bool) Issue {
	if total == 0 {
		return nil
	}

	if c.terminated > 0 && c.ready != 0 && !isInit {
		if c.reason == "" {
			return NewErrorf(WarnLevel, "Pod is terminating [%d/%d]", c.ready, total)
		}
		return NewErrorf(WarnLevel, "Pod is terminating [%d/%d] %s", c.ready, total, c.reason)
	}

	if c.terminated > 0 && c.ready == 0 {
		return nil
	}

	if c.waiting > 0 {
		if c.reason == "" {
			return NewErrorf(ErrorLevel, "Pod is waiting [%d/%d]", c.ready, total)
		}
		return NewErrorf(ErrorLevel, "Pod is waiting [%d/%d] %s", c.ready, total, c.reason)
	}

	if c.ready == 0 {
		return NewErrorf(ErrorLevel, "Pod is not ready [%d/%d]", c.ready, total)
	}

	if c.restarts > restartsLimit {
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
