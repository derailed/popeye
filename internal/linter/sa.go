package linter

import (
	"context"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

var (
	excludedSANS = []string{"kube-system"}
	excludedSAs  = []string{"default/apiproxy"}
)

// BOZO!! Check policy for potential dups or override privs

// SA represents a ServiceAccount linter.
type SA struct {
	*Linter
}

// NewSA returns a new ServiceAccount linter.
func NewSA(c Client, l *zerolog.Logger) *SA {
	return &SA{newLinter(c, l)}
}

// Lint a serviceaccount.
func (s *SA) Lint(ctx context.Context) error {
	crbs, err := s.client.ListCRBs()
	if err != nil {
		return err
	}

	rbs, err := s.client.ListRBs()
	if err != nil {
		return err
	}

	pods, err := s.client.ListAllPods()
	if err != nil {
		return nil
	}

	s.checkDead(pods, crbs, rbs)

	return nil
}

func (s *SA) checkDead(pods map[string]v1.Pod, crbs map[string]rbacv1.ClusterRoleBinding, rbs map[string]rbacv1.RoleBinding) {
	refs := make(map[string]string, len(crbs)+len(rbs))

	for _, crb := range crbs {
		pullSas(crb.Name, crb.Subjects, refs)
	}
	for _, rb := range rbs {
		pullSas(rb.Namespace+"/"+rb.Name, rb.Subjects, refs)
	}

	psas := make(map[string]struct{}, len(pods))
	for _, p := range pods {
		// Skip system namespace...
		if in(excludedSANS, p.Namespace) {
			continue
		}
		if p.Spec.ServiceAccountName != "" {
			psas[p.Namespace+"/"+p.Spec.ServiceAccountName] = struct{}{}
		}
	}

	// Check for dead service account usage
	for sa, b := range refs {
		s.initIssues(sa)
		if _, ok := psas[sa]; !ok && !in(excludedSAs, sa) {
			s.addIssuef(sa, ErrorLevel, "Used? Referenced by binding `%s", b)
		}
	}
}

func pullSas(n string, ss []rbacv1.Subject, res map[string]string) {
	for _, s := range ss {
		// Skip system namespace...
		if in(excludedSANS, s.Namespace) {
			continue
		}

		if s.Kind == "ServiceAccount" {
			fqn := fqn(s)
			if _, ok := res[fqn]; !ok {
				res[fqn] = n
			}
		}
	}
}

func fqn(s rbacv1.Subject) string {
	if s.Namespace == "" {
		return s.Name
	}
	return s.Namespace + "/" + s.Name
}
