package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
)

func TestClusterSanitize(t *testing.T) {
	uu := map[string]struct {
		major, minor string
		e            issues.Issues
	}{
		"good": {
			major: "1", minor: "15",
			e: issues.Issues{
				{Group: issues.Root, Message: "[POP-406] K8s version OK", Level: issues.OkLevel},
			},
		},
		"guizard": {
			major: "1", minor: "11",
			e: issues.Issues{
				{Group: issues.Root, Message: "[POP-405] Is this a jurassic cluster? Might want to upgrade K8s a bit", Level: issues.WarnLevel},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			cl := NewCluster(issues.NewCollector(loadCodes(t)), newCluster(u.major, u.minor))

			assert.Nil(t, cl.Sanitize(context.TODO()))
			assert.Equal(t, u.e, cl.Outcome()["Version"])
		})
	}
}

type cluster struct {
	major, minor string
}

func newCluster(major, minor string) cluster {
	return cluster{major: major, minor: minor}
}

func (c cluster) ListVersion() (string, string) {
	return c.major, c.minor
}
