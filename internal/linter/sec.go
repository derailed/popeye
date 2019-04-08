package linter

import (
	"context"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

type (
	// Sec represents a Secret linter.
	Sec struct {
		*Linter
	}

	// Reference tracks a config reference.
	Reference struct {
		name string
		keys map[string]struct{}
	}

	// TypedReferences tracks configuration type references.
	TypedReferences map[string]*Reference

	// References tracks configuration references.
	References map[string]TypedReferences
)

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
	refs := make(References, len(pods)+len(sas))

	for _, po := range pods {
		checkSecVolumes(po.Namespace, po.Spec.Volumes, refs)
		checkSecContainerRefs(po.Namespace, po.Spec.InitContainers, refs)
		checkSecContainerRefs(po.Namespace, po.Spec.Containers, refs)
		checkPullImageSecrets(po, refs)
	}

	for _, sa := range sas {
		Reference := Reference{name: sa.Name}
		for _, s := range sa.Secrets {
			fqn := fqn(sa.Namespace, s.Name)
			if v, ok := refs[fqn]; ok {
				v["sasec"] = &Reference
			} else {
				refs[fqn] = TypedReferences{"sasec": &Reference}
			}
		}

		for _, s := range sa.ImagePullSecrets {
			fqn := fqn(sa.Namespace, s.Name)
			if v, ok := refs[fqn]; ok {
				v["sapullsec"] = &Reference
			} else {
				refs[fqn] = TypedReferences{"sapullsec": &Reference}
			}
		}
	}

	for fqn, sec := range secs {
		s.initIssues(fqn)
		ref, ok := refs[fqn]
		if !ok {
			s.addIssuef(fqn, InfoLevel, "Reference?")
			continue
		}

		victims := map[string]bool{}
		for key := range sec.Data {
			victims[key] = false
			if Reference, ok := ref["volume"]; ok {
				for k := range Reference.keys {
					victims[key] = k == key
				}
			}

			if _, ok := ref["sasec"]; ok {
				victims[key] = true
			}

			if _, ok := ref["sapullsec"]; ok {
				victims[key] = true
			}

			if Reference, ok := ref["keyRef"]; ok {
				for k := range Reference.keys {
					victims[key] = k == key
				}
			}

			for k, v := range victims {
				if !v {
					s.addIssuef(fqn, InfoLevel, "Reference `%s?", k)
				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func checkPullImageSecrets(po v1.Pod, refs map[string]TypedReferences) {
	for _, s := range po.Spec.ImagePullSecrets {
		fqn := fqn(po.Namespace, s.Name)

		u := Reference{name: s.Name}
		if r, ok := refs[fqn]; ok {
			r["pull"] = &u
		} else {
			refs[fqn] = TypedReferences{"pull": &u}
		}
	}
}

func checkSecVolumes(ns string, vols []v1.Volume, refs map[string]TypedReferences) {
	for _, v := range vols {
		sec := v.VolumeSource.Secret
		if sec == nil {
			continue
		}

		fqn := fqn(ns, sec.SecretName)
		u := &Reference{name: v.Name, keys: map[string]struct{}{}}
		if r, ok := refs[fqn]; ok {
			r["volume"] = u
		} else {
			refs[fqn] = TypedReferences{"volume": u}
		}

		for _, k := range sec.Items {
			refs[fqn]["volume"].keys[k.Key] = struct{}{}
		}
	}
}

func checkSecContainerRefs(ns string, cos []v1.Container, refs map[string]TypedReferences) {
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
