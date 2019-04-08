package linter

import (
	"context"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

// CM represents a ConfigMap linter.
type CM struct {
	*Linter
}

// NewCM returns a new ConfigMap linter.
func NewCM(c Client, l *zerolog.Logger) *CM {
	return &CM{newLinter(c, l)}
}

// Lint a ConfigMap.
func (c *CM) Lint(ctx context.Context) error {
	cms, err := c.client.ListCMs()
	if err != nil {
		return err
	}

	pods, err := c.client.ListPods()
	if err != nil {
		return nil
	}

	c.lint(pods, cms)

	return nil
}

func (c *CM) lint(pods map[string]v1.Pod, cms map[string]v1.ConfigMap) {
	refs := make(References, len(cms))

	for _, po := range pods {
		checkVolumes(po.Namespace, po.Spec.Volumes, refs)
		checkContainerRefs(po.Namespace, po.Spec.InitContainers, refs)
		checkContainerRefs(po.Namespace, po.Spec.Containers, refs)
	}

	for fqn, cm := range cms {
		c.initIssues(fqn)
		ref, ok := refs[fqn]
		if !ok {
			c.addIssuef(fqn, InfoLevel, "Used?")
			continue
		}

		victims := map[string]bool{}
		for key := range cm.Data {
			victims[key] = false
			if used, ok := ref["volume"]; ok {
				if len(used.keys) != 0 {
					for k := range used.keys {
						victims[key] = k == key
					}
				} else {
					victims[key] = true
				}
			}

			if _, ok := ref["envFrom"]; ok {
				victims[key] = true
			}

			if used, ok := ref["keyRef"]; ok {
				for k := range used.keys {
					victims[key] = k == key
				}
			}

			for k, v := range victims {
				if !v {
					c.addIssuef(fqn, InfoLevel, "Used key `%s?", k)
				}
			}
		}
	}
}

func checkVolumes(ns string, vols []v1.Volume, refs References) {
	for _, v := range vols {
		cm := v.VolumeSource.ConfigMap
		if cm == nil {
			continue
		}

		fqn := fqn(ns, cm.Name)
		u := Reference{name: v.Name, keys: map[string]struct{}{}}
		if r, ok := refs[fqn]; ok {
			r["volume"] = &u
		} else {
			refs[fqn] = TypedReferences{"volume": &u}
		}

		for _, k := range cm.Items {
			refs[fqn]["volume"].keys[k.Key] = struct{}{}
		}
	}
}

func checkContainerRefs(ns string, cos []v1.Container, refs References) {
	for _, co := range cos {
		for _, e := range co.EnvFrom {
			ref := e.ConfigMapRef
			if ref == nil {
				continue
			}

			// Check EnvFrom refs...
			fqn := fqn(ns, ref.Name)
			used := Reference{}
			if v, ok := refs[fqn]; ok {
				v["envFrom"] = &used
			} else {
				refs[fqn] = TypedReferences{"envFrom": &used}
			}
		}

		// Check envs...
		for _, e := range co.Env {
			ref := e.ValueFrom
			if ref == nil {
				continue
			}

			kref := ref.ConfigMapKeyRef
			if kref == nil {
				continue
			}

			fqn := fqn(ns, kref.Name)
			if v, ok := refs[fqn]; ok {
				v["keyRef"].keys[kref.Name] = struct{}{}
			} else {
				refs[fqn] = map[string]*Reference{
					"keyRef": &Reference{
						name: kref.Name,
						keys: map[string]struct{}{
							kref.Key: struct{}{},
						},
					},
				}
			}
		}
	}
}

func fqnCM(s v1.ConfigMap) string {
	return s.Namespace + "/" + s.Name
}

func fqn(ns, n string) string {
	return ns + "/" + n
}
