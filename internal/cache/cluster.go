// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import "github.com/Masterminds/semver"

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
func (c *Cluster) ListVersion() *semver.Version {
	return c.rev
}
