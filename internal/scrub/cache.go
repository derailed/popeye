package scrub

import (
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
)

type dial struct {
	factory types.Factory
	config  *config.Config
}

func newDial(f types.Factory, cfg *config.Config) *dial {
	return &dial{
		factory: f,
		config:  cfg,
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
func NewCache(f types.Factory, cfg *config.Config) *Cache {
	d := newDial(f, cfg)
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
