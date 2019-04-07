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

type usedCM struct {
	name string
	keys map[string]struct{}
}
type usedCMs map[string]*usedCM

func (c *CM) lint(pods map[string]v1.Pod, cms map[string]v1.ConfigMap) {
	refs := make(map[string]usedCMs, len(cms))

	for _, po := range pods {
		checkVolumes(po.Namespace, po.Spec.Volumes, refs)
		checkContainerRefs(po.Namespace, po.Spec.InitContainers, refs)
		checkContainerRefs(po.Namespace, po.Spec.Containers, refs)
	}

	for fqn, cm := range cms {
		c.initIssues(fqn)
		ref, ok := refs[fqn]
		if !ok {
			c.addIssuef(fqn, WarnLevel, "Used?")
			continue
		}

		victims := map[string]bool{}
		for key := range cm.Data {
			victims[key] = false
			if used, ok := ref["volume"]; ok {
				for k := range used.keys {
					victims[key] = k == key
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
					c.addIssuef(fqn, WarnLevel, "Found unused key `%s", k)
				}
			}
		}
	}
}

func checkVolumes(ns string, vols []v1.Volume, refs map[string]usedCMs) {
	for _, v := range vols {
		cm := v.VolumeSource.ConfigMap
		if cm == nil {
			continue
		}

		fqn := fqn(ns, cm.Name)
		u := &usedCM{name: v.Name, keys: map[string]struct{}{}}
		if r, ok := refs[fqn]; ok {
			r["volume"] = u
		} else {
			refs[fqn] = usedCMs{"volume": u}
		}

		for _, k := range cm.Items {
			refs[fqn]["volume"].keys[k.Key] = struct{}{}
		}
	}
}

func checkContainerRefs(ns string, cos []v1.Container, refs map[string]usedCMs) {
	for _, co := range cos {
		for _, e := range co.EnvFrom {
			ref := e.ConfigMapRef
			if ref == nil {
				continue
			}

			// Check EnvFrom refs...
			fqn := fqn(ns, ref.Name)
			used := usedCM{}
			if v, ok := refs[fqn]; ok {
				v["envFrom"] = &used
			} else {
				refs[fqn] = map[string]*usedCM{"envFrom": &used}
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
				refs[fqn] = map[string]*usedCM{
					"keyRef": &usedCM{
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
