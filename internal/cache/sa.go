package cache

import (
	"github.com/derailed/popeye/internal"
	v1 "k8s.io/api/core/v1"
)

// ServiceAccount tracks serviceaccounts.
type ServiceAccount struct {
	sas map[string]*v1.ServiceAccount
}

// NewServiceAccount returns a new serviceaccount loader.
func NewServiceAccount(sas map[string]*v1.ServiceAccount) *ServiceAccount {
	return &ServiceAccount{sas: sas}
}

// ListServiceAccounts list available ServiceAccounts.
func (s *ServiceAccount) ListServiceAccounts() map[string]*v1.ServiceAccount {
	return s.sas
}

// ServiceAccountRefs computes all serviceaccount external references.
func (s *ServiceAccount) ServiceAccountRefs(refs ObjReferences) {
	for _, sa := range s.sas {
		for _, s := range sa.Secrets {
			key := ResFqn(SecretKey, FQN(s.Namespace, s.Name))
			if set, ok := refs[key]; ok {
				set.Add(AllKeys)
			} else {
				refs[key] = internal.StringSet{AllKeys: internal.Blank}
			}
		}

		for _, s := range sa.ImagePullSecrets {
			key := ResFqn(SecretKey, FQN(sa.Namespace, s.Name))
			if set, ok := refs[key]; ok {
				set.Add(AllKeys)
			} else {
				refs[key] = internal.StringSet{AllKeys: internal.Blank}
			}
		}
	}
}
