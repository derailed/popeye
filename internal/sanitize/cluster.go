package sanitize

import (
	"context"
	"strconv"

	"github.com/derailed/popeye/internal/issues"
)

const (
	tolerableMajor = 1
	tolerableMinor = 12
)

type (
	// Cluster tracks cluster sanitization.
	Cluster struct {
		*issues.Collector
		ClusterLister
	}

	// ClusterLister list available Clusters on a cluster.
	ClusterLister interface {
		ListVersion() (string, string)
	}
)

// NewCluster returns a new Cluster sanitizer.
func NewCluster(co *issues.Collector, lister ClusterLister) *Cluster {
	return &Cluster{
		Collector:     co,
		ClusterLister: lister,
	}
}

// Sanitize configmaps.
func (c *Cluster) Sanitize(ctx context.Context) error {
	major, minor := c.ListVersion()

	m, err := strconv.Atoi(major)
	if err != nil {
		return err
	}
	p, err := strconv.Atoi(minor)
	if err != nil {
		return err
	}

	if m != tolerableMajor || p < tolerableMinor {
		c.AddCode(405, "Version")
	} else {
		c.AddCode(406, "Version")
	}

	return nil
}
