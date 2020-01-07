package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
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

	// IngressRefs tracks Ingress object references.
	IngressRefs interface {
		IngressRefs(cache.ObjReferences)
	}

	// SecretLister list available Secrets on a cluster.
	SecretLister interface {
		PodRefs
		SARefs
		IngressRefs
		ListSecrets() map[string]*v1.Secret
	}
)

// NewSecret returns a new sanitizer.
func NewSecret(co *issues.Collector, lister SecretLister) *Secret {
	return &Secret{
		Collector:    co,
		SecretLister: lister,
	}
}

// Sanitize cleanse the resource.
func (s *Secret) Sanitize(ctx context.Context) error {
	refs := cache.ObjReferences{}

	s.PodRefs(refs)
	s.ServiceAccountRefs(refs)
	s.IngressRefs(refs)
	s.checkInUse(ctx, refs)

	return nil
}

func (s *Secret) checkInUse(ctx context.Context, refs cache.ObjReferences) {
	for fqn, sec := range s.ListSecrets() {
		s.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)
		defer func(fqn string, ctx context.Context) {
			if s.Config.ExcludeFQN(internal.MustExtractSection(ctx), fqn) {
				s.ClearOutcome(fqn)
			}
		}(fqn, ctx)

		keys, ok := refs[cache.ResFqn(cache.SecretKey, fqn)]
		if !ok {
			s.AddCode(ctx, 400)
			continue
		}
		if keys.Has(cache.AllKeys) {
			continue
		}

		kk := make(internal.StringSet, len(sec.Data))
		for k := range sec.Data {
			kk.Add(k)
		}
		deltas := keys.Diff(kk)
		for k := range deltas {
			s.AddCode(ctx, 401, k)
		}
	}
}
