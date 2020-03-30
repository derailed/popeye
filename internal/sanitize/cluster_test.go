package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
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
				{
					GVR:     "clusters",
					Group:   issues.Root,
					Message: "[POP-406] K8s version OK",
					Level:   config.OkLevel,
				},
			},
		},
		"guizard": {
			major: "1", minor: "11",
			e: issues.Issues{
				{
					GVR:     "clusters",
					Group:   issues.Root,
					Message: "[POP-405] Is this a jurassic cluster? Might want to upgrade K8s a bit",
					Level:   config.WarnLevel,
				},
			},
		},
	}

	ctx := makeContext("clusters", "cluster")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cl := NewCluster(issues.NewCollector(loadCodes(t), makeConfig(t)), newCluster(u.major, u.minor))

			assert.Nil(t, cl.Sanitize(ctx))
			assert.Equal(t, u.e, cl.Outcome()["Version"])
		})
	}
}

// Helpers...

func makeConfig(t *testing.T) *config.Config {
	c, err := config.NewConfig(config.NewFlags())
	assert.Nil(t, err)
	return c
}

func makeContext(gvr, section string) context.Context {
	return context.WithValue(context.Background(), internal.KeyRunInfo, internal.RunInfo{
		Section:    section,
		SectionGVR: client.NewGVR(gvr),
	})
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
