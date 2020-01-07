package cache

import (
	v1 "k8s.io/api/core/v1"
)

// SecretKey tracks Secret resource references
const SecretKey = "sec"

// Secret represents a collection of Secrets available on a cluster.
type Secret struct {
	secrets map[string]*v1.Secret
}

// NewSecret returns a new Secret cache.
func NewSecret(ss map[string]*v1.Secret) *Secret {
	return &Secret{ss}
}

// ListSecrets returns all available Secrets on the cluster.
func (s *Secret) ListSecrets() map[string]*v1.Secret {
	return s.secrets
}
