package cache

// ClusterKey tracks Cluster resource references
const ClusterKey = "cl"

// Cluster represents Cluster cache.
type Cluster struct {
	major, minor string
}

// NewCluster returns a new Cluster cache.
func NewCluster(major, minor string) *Cluster {
	return &Cluster{major: major, minor: minor}
}

// ListVersion returns cluster server version.
func (c *Cluster) ListVersion() (string, string) {
	return c.major, c.minor
}
