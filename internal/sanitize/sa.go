package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// BOZO!! Check policy for potential dups or override priviledges

type (
	// PodLister lists available pods.
	PodLister interface {
		ListPods() map[string]*v1.Pod
		GetPod(sel map[string]string) *v1.Pod
	}

	// ServiceAccountLister list available ServiceAccounts on a cluster.
	ServiceAccountLister interface {
		PodLister
		ClusterRoleBindingLister
		RoleBindingLister
		SecretLister

		ListServiceAccounts() map[string]*v1.ServiceAccount
	}

	// ClusterRoleBindingLister list all available ClusterRoleBindings.
	ClusterRoleBindingLister interface {
		ListClusterRoleBindings() map[string]*rbacv1.ClusterRoleBinding
	}

	// RoleBindingLister list all available ClusterRoleBindings.
	RoleBindingLister interface {
		ListRoleBindings() map[string]*rbacv1.RoleBinding
	}

	// ServiceAccount tracks ServiceAccount sanitizer.
	ServiceAccount struct {
		*issues.Collector

		ServiceAccountLister
	}
)

// NewServiceAccount returns a new ServiceAccount linter.
func NewServiceAccount(co *issues.Collector, lister ServiceAccountLister) *ServiceAccount {
	return &ServiceAccount{
		Collector:            co,
		ServiceAccountLister: lister,
	}

}

// Sanitize a serviceaccount.
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
		s.checkMounts(fqn, sa.AutomountServiceAccountToken)
		s.checkSecretRefs(fqn, sa.Secrets)
		s.checkPullSecretRefs(fqn, sa.ImagePullSecrets)
		if _, ok := refs[fqn]; !ok {
			s.AddCode(400, fqn)
		}
	}

	return nil
}

func (s *ServiceAccount) checkSecretRefs(fqn string, refs []v1.ObjectReference) {
	ns, _ := namespaced(fqn)
	for _, ref := range refs {
		if ref.Namespace != "" {
			ns = ref.Namespace
		}
		sfqn := cache.FQN(ns, ref.Name)
		if _, ok := s.ListSecrets()[sfqn]; !ok {
			s.AddCode(304, fqn, sfqn)
		}
	}
}

func (s *ServiceAccount) checkPullSecretRefs(fqn string, refs []v1.LocalObjectReference) {
	ns, _ := namespaced(fqn)
	for _, ref := range refs {
		sfqn := cache.FQN(ns, ref.Name)
		if _, ok := s.ListSecrets()[sfqn]; !ok {
			s.AddCode(305, fqn, sfqn)
		}
	}
}

func (s *ServiceAccount) checkMounts(fqn string, b *bool) {
	if b != nil && *b {
		s.AddCode(303, fqn)
	}
}

func (s *ServiceAccount) crbRefs(refs map[string]struct{}) error {
	for _, crb := range s.ListClusterRoleBindings() {
		pullSas(crb.Name, crb.Subjects, refs)
	}

	return nil
}

func (s *ServiceAccount) rbRefs(refs map[string]struct{}) error {
	for _, rb := range s.ListRoleBindings() {
		pullSas(cache.FQN(rb.Namespace, rb.Name), rb.Subjects, refs)
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

func pullSas(n string, ss []rbacv1.Subject, res map[string]struct{}) {
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
