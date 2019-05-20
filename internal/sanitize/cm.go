package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

type (
	// ConfigMap tracks ConfigMap sanitization.
	ConfigMap struct {
		*issues.Collector
		ConfigMapLister
	}

	// PodRefs tracks pods object references.
	PodRefs interface {
		PodRefs(cache.ObjReferences)
	}

	// ConfigMapLister list available ConfigMaps on a cluster.
	ConfigMapLister interface {
		PodRefs
		ListConfigMaps() map[string]*v1.ConfigMap
	}
)

// NewConfigMap returns a new ConfigMap sanitizer.
func NewConfigMap(c *issues.Collector, lister ConfigMapLister) *ConfigMap {
	return &ConfigMap{
		Collector:       c,
		ConfigMapLister: lister,
	}
}

// Sanitize a configmap.
func (c *ConfigMap) Sanitize(context.Context) error {
	cmRefs := cache.ObjReferences{}
	c.PodRefs(cmRefs)
	c.checkInUse(cmRefs)

	return nil
}

func (c *ConfigMap) checkInUse(refs cache.ObjReferences) {
	for fqn, cm := range c.ListConfigMaps() {
		c.InitOutcome(fqn)

		keys, ok := refs[cache.ResFqn(cache.ConfigMapKey, fqn)]
		if !ok {
			c.AddInfo(fqn, "Used?")
			continue
		}
		if keys.Has(cache.AllKeys) {
			continue
		}

		kk := make(cache.StringSet, len(cm.Data))
		for k := range cm.Data {
			kk.Add(k)
		}
		deltas := keys.Diff(kk)
		if len(deltas) == 0 {
			continue
		}
		for k := range deltas {
			c.AddInfof(fqn, "Key `%s` might not be used?", k)
		}
	}
}
