package scrub

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
)

type dial struct {
	client *k8s.Client
	config *config.Config
}

func newDial(c *k8s.Client, cfg *config.Config) *dial {
	return &dial{
		client: c,
		config: cfg,
	}
}

// Cache tracks commonly used resources.
type Cache struct {
	*dial
	*core
	*apps
	*rbac
	*policy
	*ext
	*mx
}

// NewCache returns a new resource cache
func NewCache(c *k8s.Client, cfg *config.Config) *Cache {
	d := newDial(c, cfg)
	return &Cache{
		dial:   d,
		core:   newCore(d),
		apps:   newApps(d),
		rbac:   newRBAC(d),
		policy: newPolicy(d),
		ext:    newExt(d),
		mx:     newMX(d),
	}
}
