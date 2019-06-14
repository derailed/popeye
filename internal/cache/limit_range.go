package cache

import (
	v1 "k8s.io/api/core/v1"
)

// LimitRangeKey tracks LimitRange resource references
const LimitRangeKey = "lr"

// LimitRange represents LimitRange cache.
type LimitRange struct {
	lrs map[string]*v1.LimitRange
}

// NewLimitRange returns a new LimitRange cache.
func NewLimitRange(lrs map[string]*v1.LimitRange) *LimitRange {
	return &LimitRange{lrs: lrs}
}

// ListLimitRanges returns all available LimitRanges on the cluster.
func (c *LimitRange) ListLimitRanges() map[string]*v1.LimitRange {
	return c.lrs
}
