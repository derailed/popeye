package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
)

// ListVersion return server api version.
func ListVersion(c *k8s.Client, cfg *config.Config) (string, string, error) {
	v, err := c.DialOrDie().Discovery().ServerVersion()
	if err != nil {
		return "", "", err
	}

	return v.Major, v.Minor, nil
}
