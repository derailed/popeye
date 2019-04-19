package linter

import (
	"context"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

// ConfigMap represents a ConfigMap linter.
type ConfigMap struct {
	*Linter
}

// NewConfigMap returns a new ConfigMap linter.
func NewConfigMap(l Loader, log *zerolog.Logger) *ConfigMap {
	return &ConfigMap{NewLinter(l, log)}
}

// Lint a ConfigMap.
func (c *ConfigMap) Lint(ctx context.Context) error {
	cms, err := c.ListConfigMaps()
	if err != nil {
		return err
	}

	pods, err := c.ListPods()
	if err != nil {
		return nil
	}

	c.lint(cms, pods)

	return nil
}

func (c *ConfigMap) lint(cms map[string]v1.ConfigMap, pods map[string]v1.Pod) {
	refs := make(References, len(cms))
	for fqn, po := range pods {
		c.checkVolumes(fqn, po.Spec.Volumes, refs)
		c.checkContainerRefs(fqn, po.Spec.InitContainers, refs)
		c.checkContainerRefs(fqn, po.Spec.Containers, refs)
	}

	for fqn, cm := range cms {
		c.initIssues(fqn)
		cmRef, ok := refs[fqn]
		if !ok {
			c.addIssuef(fqn, InfoLevel, "Used?")
			continue
		}

		victims := map[string]bool{}
		for key := range cm.Data {
			victims[key] = false
			if used, ok := cmRef["volume"]; ok {
				// If volumes does not specify items, then all cm keys used!
				if len(used.keys) == 0 {
					victims[key] = true
					continue
				}
				if _, ok := used.keys[key]; ok {
					victims[key] = true
					continue
				}
			}

			if _, ok := cmRef["envFrom"]; ok {
				victims[key] = true
				continue
			}

			if used, ok := cmRef["env"]; ok {
				if _, ok := used.keys[key]; ok {
					victims[key] = true
				}
			}
		}

		for k, v := range victims {
			if v {
				continue
			}
			c.addIssuef(fqn, InfoLevel, "Unused key `%s?", k)
		}
	}
}

func (*ConfigMap) checkVolumes(poFQN string, vols []v1.Volume, refs References) {
	ns, _ := namespaced(poFQN)
	for _, v := range vols {
		cm := v.VolumeSource.ConfigMap
		if cm == nil {
			continue
		}

		key := fqn(ns, cm.Name)
		u := Reference{name: ref(poFQN, v.Name), keys: map[string]struct{}{}}
		if r, ok := refs[key]; ok {
			r["volume"] = &u
		} else {
			refs[key] = TypedReferences{"volume": &u}
		}

		for _, k := range cm.Items {
			refs[key]["volume"].keys[k.Key] = struct{}{}
		}
	}
}

func (c *ConfigMap) checkContainerRefs(poFQN string, cos []v1.Container, refs References) {
	for _, co := range cos {
		c.checkEnvFrom(poFQN, co, refs)
		c.checkEnv(poFQN, co, refs)
	}
}

// CheckEnvFrom check container envFrom for configMap references.
func (*ConfigMap) checkEnvFrom(poFQN string, co v1.Container, refs References) {
	ns, _ := namespaced(poFQN)
	for _, e := range co.EnvFrom {
		cmRef := e.ConfigMapRef
		if cmRef == nil {
			continue
		}

		key := fqn(ns, cmRef.Name)
		used := Reference{name: ref(poFQN, co.Name)}
		if v, ok := refs[key]; ok {
			v["envFrom"] = &used
			continue
		}
		refs[key] = TypedReferences{"envFrom": &used}
	}
}

// CheckEnv checks container Env section for configMap references.
func (*ConfigMap) checkEnv(poFQN string, co v1.Container, refs References) {
	ns, _ := namespaced(poFQN)
	blank := struct{}{}

	for _, e := range co.Env {
		if e.ValueFrom == nil || e.ValueFrom.ConfigMapKeyRef == nil {
			continue
		}

		kref := e.ValueFrom.ConfigMapKeyRef
		key := fqn(ns, kref.Name)
		if v, ok := refs[key]; ok {
			if kv, ok := v["env"]; ok {
				kv.keys[kref.Name] = blank
			} else {
				v["env"] = &Reference{
					name: kref.Name,
					keys: map[string]struct{}{
						kref.Key: blank,
					},
				}
			}
			continue
		}
		refs[key] = TypedReferences{
			"env": {
				name: kref.Name,
				keys: map[string]struct{}{
					kref.Key: blank,
				},
			},
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func ref(s, r string) string {
	return s + ":" + r
}

func fqnConfigMap(s v1.ConfigMap) string {
	return s.Namespace + "/" + s.Name
}
