package sanitize

import (
	"context"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// BOZO!! Check policy for potential dups or override priviledges

type (
	// ServiceAccountLister list available ServiceAccounts on a cluster.
	ServiceAccountLister interface {
		PodLister
		ClusterRoleBindingLister
		RoleBindingLister
		SecretLister

		ListServiceAccounts() map[string]*v1.ServiceAccount
	}

	// ClusterRoleBindingRefs tracks crb references.
	ClusterRoleBindingRefs interface {
		ClusterRoleRefs(*sync.Map)
	}

	// RoleBindingRefs tracks rb references.
	RoleBindingRefs interface {
		RoleRefs(*sync.Map)
	}

	// ClusterRoleBindingLister list all available ClusterRoleBindings.
	ClusterRoleBindingLister interface {
		ClusterRoleBindingRefs
		ListClusterRoleBindings() map[string]*rbacv1.ClusterRoleBinding
	}

	// RoleBindingLister list all available ClusterRoleBindings.
	RoleBindingLister interface {
		RoleBindingRefs
		ListRoleBindings() map[string]*rbacv1.RoleBinding
	}

	// ServiceAccount tracks ServiceAccount sanitizer.
	ServiceAccount struct {
		*issues.Collector

		ServiceAccountLister
	}
)

// NewServiceAccount returns a new sanitizer.
func NewServiceAccount(co *issues.Collector, lister ServiceAccountLister) *ServiceAccount {
	return &ServiceAccount{
		Collector:            co,
		ServiceAccountLister: lister,
	}

}

// Sanitize cleanse the resource.
func (s *ServiceAccount) Sanitize(ctx context.Context) error {
	refs := make(map[string]struct{}, 20)
	if err := s.crbRefs(refs); err != nil {
		return err
	}
	if err := s.rbRefs(refs); err != nil {
		return err
	}
	err := s.podRefs(refs)
	if err != nil {
		return err
	}

	for fqn, sa := range s.ListServiceAccounts() {
		s.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		s.checkMounts(ctx, sa.AutomountServiceAccountToken)
		s.checkSecretRefs(ctx, sa.Secrets)
		s.checkPullSecretRefs(ctx, sa.ImagePullSecrets)
		if _, ok := refs[fqn]; !ok {
			s.AddCode(ctx, 400)
		}

		if s.NoConcerns(fqn) && s.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			s.ClearOutcome(fqn)
		}
	}

	return nil
}

func (s *ServiceAccount) checkSecretRefs(ctx context.Context, refs []v1.ObjectReference) {
	ns, _ := namespaced(internal.MustExtractFQN(ctx))
	for _, ref := range refs {
		if ref.Namespace != "" {
			ns = ref.Namespace
		}
		sfqn := cache.FQN(ns, ref.Name)
		if _, ok := s.ListSecrets()[sfqn]; !ok {
			s.AddCode(ctx, 304, sfqn)
		}
	}
}

func (s *ServiceAccount) checkPullSecretRefs(ctx context.Context, refs []v1.LocalObjectReference) {
	ns, _ := namespaced(internal.MustExtractFQN(ctx))
	for _, ref := range refs {
		sfqn := cache.FQN(ns, ref.Name)
		if _, ok := s.ListSecrets()[sfqn]; !ok {
			s.AddCode(ctx, 305, sfqn)
		}
	}
}

func (s *ServiceAccount) checkMounts(ctx context.Context, b *bool) {
	if b != nil && *b {
		s.AddCode(ctx, 303)
	}
}

func (s *ServiceAccount) crbRefs(refs map[string]struct{}) error {
	for _, crb := range s.ListClusterRoleBindings() {
		pullSas(crb.Subjects, refs)
	}

	return nil
}

func (s *ServiceAccount) rbRefs(refs map[string]struct{}) error {
	for _, rb := range s.ListRoleBindings() {
		pullSas(rb.Subjects, refs)
	}

	return nil
}

func (s *ServiceAccount) podRefs(refs map[string]struct{}) error {
	for _, p := range s.ListPods() {
		if p.Spec.ServiceAccountName != "" {
			refs[cache.FQN(p.Namespace, p.Spec.ServiceAccountName)] = struct{}{}
		}
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func pullSas(ss []rbacv1.Subject, res map[string]struct{}) {
	for _, s := range ss {
		if s.Kind == "ServiceAccount" {
			fqn := fqnSubject(s)
			if _, ok := res[fqn]; !ok {
				res[fqn] = struct{}{}
			}
		}
	}
}

func fqnSubject(s rbacv1.Subject) string {
	return cache.FQN(s.Namespace, s.Name)
}
