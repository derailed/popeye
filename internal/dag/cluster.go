package dag

import (
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
)

// ListVersion return server api version.
func ListVersion(c types.Connection, cfg *config.Config) (string, string, error) {
	v, err := c.DialOrDie().Discovery().ServerVersion()
	if err != nil {
		return "", "", err
	}

	return v.Major, v.Minor, nil
}
