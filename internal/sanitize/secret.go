package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

type (
	// Secret tracks Secret sanitization.
	Secret struct {
		*issues.Collector
		SecretLister
	}

	// SARefs tracks ServiceAccount object references.
	SARefs interface {
		ServiceAccountRefs(cache.ObjReferences)
	}

	// SecretLister list available Secrets on a cluster.
	SecretLister interface {
		PodRefs
		SARefs
		ListSecrets() map[string]*v1.Secret
	}
)

// NewSecret returns a new Secret sanitizer.
func NewSecret(co *issues.Collector, lister SecretLister) *Secret {
	return &Secret{
		Collector:    co,
		SecretLister: lister,
	}
}

// Sanitize a secret.
func (s *Secret) Sanitize(context.Context) error {
	refs := cache.ObjReferences{}

	s.PodRefs(refs)
	s.ServiceAccountRefs(refs)
	s.checkInUse(refs)

	return nil
}

func (s *Secret) checkInUse(refs cache.ObjReferences) {
	for fqn, sec := range s.ListSecrets() {
		s.InitOutcome(fqn)

		keys, ok := refs[cache.ResFqn(cache.SecretKey, fqn)]
		if !ok {
			s.AddInfo(fqn, "Used?")
			continue
		}
		if keys.Has(cache.AllKeys) {
			continue
		}

		kk := make(cache.StringSet, len(sec.Data))
		for k := range sec.Data {
			kk.Add(k)
		}
		deltas := keys.Diff(kk)
		if len(deltas) == 0 {
			continue
		}
		for k := range deltas {
			s.AddInfof(fqn, "Key `%s` might not be used?", k)
		}
	}
}
