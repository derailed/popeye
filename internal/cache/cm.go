package cache

import (
	v1 "k8s.io/api/core/v1"
)

// ConfigMapKey tracks ConfigMap resource references
const ConfigMapKey = "cm"

// ConfigMap represents ConfigMap cache.
type ConfigMap struct {
	cms map[string]*v1.ConfigMap
}

// NewConfigMap returns a new ConfigMap cache.
func NewConfigMap(cms map[string]*v1.ConfigMap) *ConfigMap {
	return &ConfigMap{cms: cms}
}

// ListConfigMaps returns all available ConfigMaps on the cluster.
func (c *ConfigMap) ListConfigMaps() map[string]*v1.ConfigMap {
	return c.cms
}
