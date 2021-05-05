package sanitize

import (
	"context"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

type (
	// PodRefs tracks pods object references.
	PodRefs interface {
		PodRefs(*sync.Map)
	}

	// ConfigMapLister list available ConfigMaps on a cluster.
	ConfigMapLister interface {
		PodRefs
		ListConfigMaps() map[string]*v1.ConfigMap
	}

	// ConfigMap tracks ConfigMap sanitization.
	ConfigMap struct {
		*issues.Collector
		ConfigMapLister
	}
)

// NewConfigMap returns a new sanitizer.
func NewConfigMap(c *issues.Collector, lister ConfigMapLister) *ConfigMap {
	return &ConfigMap{
		Collector:       c,
		ConfigMapLister: lister,
	}
}

// Sanitize cleanse the resource.
func (c *ConfigMap) Sanitize(ctx context.Context) error {
	var cmRefs sync.Map
	c.PodRefs(&cmRefs)
	c.checkInUse(ctx, &cmRefs)

	return nil
}

func (c *ConfigMap) checkInUse(ctx context.Context, refs *sync.Map) {
	for fqn, cm := range c.ListConfigMaps() {
		c.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)
		keys, ok := refs.Load(cache.ResFqn(cache.ConfigMapKey, fqn))
		defer func(ctx context.Context, fqn string) {
			if c.NoConcerns(fqn) && c.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
				c.ClearOutcome(fqn)
			}
		}(ctx, fqn)
		if !ok {
			c.AddCode(ctx, 400)
			continue
		}
		if keys.(internal.StringSet).Has(internal.All) {
			continue
		}

		kk := make(internal.StringSet, len(cm.Data))
		for k := range cm.Data {
			kk.Add(k)
		}
		deltas := keys.(internal.StringSet).Diff(kk)
		for k := range deltas {
			c.AddCode(ctx, 401, k)
		}
	}
}
