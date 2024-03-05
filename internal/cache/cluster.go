// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"errors"

	"github.com/blang/semver/v4"
)

// ClusterKey tracks Cluster resource references
const ClusterKey = "cl"

// Cluster represents Cluster cache.
type Cluster struct {
	rev *semver.Version
}

// NewCluster returns a new Cluster cache.
func NewCluster(v *semver.Version) *Cluster {
	return &Cluster{rev: v}
}

// ListVersion returns cluster server version.
func (c *Cluster) ListVersion() (*semver.Version, error) {
	if c.rev == nil {
		return nil, errors.New("unable to assert cluster version")
	}

	return c.rev, nil
}
