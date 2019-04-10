package linter

import (
	"context"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

type (
	// Secret represents a Secret linter.
	Secret struct {
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

// NewSecret returns a new Secret linter.
func NewSecret(l Loader, log *zerolog.Logger) *Secret {
	return &Secret{NewLinter(l, log)}
}

// Lint a Secret.
func (s *Secret) Lint(ctx context.Context) error {
	secs, err := s.ListSecs()
	if err != nil {
		return err
	}

	pods, err := s.ListPods()
	if err != nil {
		return nil
	}

	sas, err := s.ListSAs()
	if err != nil {
		return nil
	}

	s.lint(secs, pods, sas)

	return nil
}

func (s *Secret) lint(secs map[string]v1.Secret, pods map[string]v1.Pod, sas map[string]v1.ServiceAccount) {
	refs := make(References, len(pods)+len(sas))

	for fqn, po := range pods {
		s.checkVolumes(fqn, po.Spec.Volumes, refs)
		s.checkContainerRefs(fqn, po.Spec.InitContainers, refs)
		s.checkContainerRefs(fqn, po.Spec.Containers, refs)
		s.checkPullImageSecrets(po, refs)
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

			if Reference, ok := ref["env"]; ok {
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

func (*Secret) checkPullImageSecrets(po v1.Pod, refs map[string]TypedReferences) {
	for _, s := range po.Spec.ImagePullSecrets {
		fqn := fqn(po.Namespace, s.Name)

		u := Reference{name: s.Name}
		if r, ok := refs[fqn]; ok {
			r["pull"] = &u
			continue
		}
		refs[fqn] = TypedReferences{"pull": &u}
	}
}

func (*Secret) checkVolumes(poFQN string, vols []v1.Volume, refs map[string]TypedReferences) {
	ns, _ := namespaced(poFQN)
	for _, v := range vols {
		if v.VolumeSource.Secret == nil {
			continue
		}

		sec := v.VolumeSource.Secret
		fqn := fqn(ns, sec.SecretName)
		u := &Reference{name: ref(poFQN, v.Name), keys: map[string]struct{}{}}
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

func (*Secret) checkContainerRefs(poFQN string, cos []v1.Container, refs map[string]TypedReferences) {
	ns, _ := namespaced(poFQN)
	for _, co := range cos {
		// Check envs...
		for _, e := range co.Env {
			if e.ValueFrom == nil || e.ValueFrom.SecretKeyRef == nil {
				continue
			}

			kref := e.ValueFrom.SecretKeyRef
			fqn := fqn(ns, kref.Name)
			if v, ok := refs[fqn]; ok {
				v["env"].keys[kref.Name] = struct{}{}
				continue
			}

			refs[fqn] = map[string]*Reference{
				"env": &Reference{
					name: kref.Name,
					keys: map[string]struct{}{
						kref.Key: struct{}{},
					},
				},
			}
		}
	}
}
