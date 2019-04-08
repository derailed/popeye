package linter

import (
	"context"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

// Sec represents a Secret linter.
type Sec struct {
	*Linter
}

type usedSec struct {
	name string
	keys map[string]struct{}
}

type usedSecs map[string]*usedSec

// NewSec returns a new Secret linter.
func NewSec(c Client, l *zerolog.Logger) *Sec {
	return &Sec{newLinter(c, l)}
}

// Lint a Secret.
func (s *Sec) Lint(ctx context.Context) error {
	secs, err := s.client.ListSecs()
	if err != nil {
		return err
	}

	pods, err := s.client.ListPods()
	if err != nil {
		return nil
	}

	sas, err := s.client.ListSAs()
	if err != nil {
		return nil
	}

	s.lint(secs, pods, sas)

	return nil
}

func (s *Sec) lint(secs map[string]v1.Secret, pods map[string]v1.Pod, sas map[string]v1.ServiceAccount) {
	refs := make(map[string]usedSecs, len(pods)+len(sas))

	for _, po := range pods {
		checkSecVolumes(po.Namespace, po.Spec.Volumes, refs)
		checkSecContainerRefs(po.Namespace, po.Spec.InitContainers, refs)
		checkSecContainerRefs(po.Namespace, po.Spec.Containers, refs)
		checkPullImageSecrets(po, refs)
	}

	for _, sa := range sas {
		used := usedSec{name: sa.Name}
		for _, s := range sa.Secrets {
			fqn := fqn(sa.Namespace, s.Name)
			if v, ok := refs[fqn]; ok {
				v["sasec"] = &used
			} else {
				refs[fqn] = usedSecs{"sasec": &used}
			}
		}

		for _, s := range sa.ImagePullSecrets {
			fqn := fqn(sa.Namespace, s.Name)
			if v, ok := refs[fqn]; ok {
				v["sapullsec"] = &used
			} else {
				refs[fqn] = usedSecs{"sapullsec": &used}
			}
		}
	}

	for fqn, sec := range secs {
		s.initIssues(fqn)
		ref, ok := refs[fqn]
		if !ok {
			s.addIssuef(fqn, InfoLevel, "Used?")
			continue
		}

		victims := map[string]bool{}
		for key := range sec.Data {
			victims[key] = false
			if used, ok := ref["volume"]; ok {
				for k := range used.keys {
					victims[key] = k == key
				}
			}

			if _, ok := ref["sasec"]; ok {
				victims[key] = true
			}

			if _, ok := ref["sapullsec"]; ok {
				victims[key] = true
			}

			if used, ok := ref["keyRef"]; ok {
				for k := range used.keys {
					victims[key] = k == key
				}
			}

			for k, v := range victims {
				if !v {
					s.addIssuef(fqn, InfoLevel, "Found unused key `%s", k)
				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func checkPullImageSecrets(po v1.Pod, refs map[string]usedSecs) {
	for _, s := range po.Spec.ImagePullSecrets {
		fqn := fqn(po.Namespace, s.Name)
		u := &usedSec{name: s.Name}
		if r, ok := refs[fqn]; ok {
			r["pull"] = u
		} else {
			refs[fqn] = usedSecs{"pull": u}
		}
	}
}

func checkSecVolumes(ns string, vols []v1.Volume, refs map[string]usedSecs) {
	for _, v := range vols {
		sec := v.VolumeSource.Secret
		if sec == nil {
			continue
		}

		fqn := fqn(ns, sec.SecretName)
		u := &usedSec{name: v.Name, keys: map[string]struct{}{}}
		if r, ok := refs[fqn]; ok {
			r["volume"] = u
		} else {
			refs[fqn] = usedSecs{"volume": u}
		}

		for _, k := range sec.Items {
			refs[fqn]["volume"].keys[k.Key] = struct{}{}
		}
	}
}

func checkSecContainerRefs(ns string, cos []v1.Container, refs map[string]usedSecs) {
	for _, co := range cos {
		// Check envs...
		for _, e := range co.Env {
			ref := e.ValueFrom
			if ref == nil {
				continue
			}

			kref := ref.SecretKeyRef
			if kref == nil {
				continue
			}

			fqn := fqn(ns, kref.Name)
			if v, ok := refs[fqn]; ok {
				v["keyRef"].keys[kref.Name] = struct{}{}
			} else {
				refs[fqn] = map[string]*usedSec{
					"keyRef": &usedSec{
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
