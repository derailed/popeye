package linter

import (
	"context"
	"strings"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
	crbs := map[string]rbacv1.ClusterRoleBinding{}
	if s.client.ActiveNamespace() == "" {
		var err error
		crbs, err = s.client.ListAllCRBs()
		if err != nil {
			return err
		}
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
		if s.client.ExcludedNS(rb.Namespace) {
			continue
		}
		pullSas(rb.Namespace+"/"+rb.Name, rb.Subjects, refs)
	}

	psas := make(map[string]struct{}, len(pods))
	for _, p := range pods {
		if s.client.ExcludedNS(p.Namespace) {
			continue
		}

		if p.Spec.ServiceAccountName != "" {
			psas[p.Namespace+"/"+p.Spec.ServiceAccountName] = struct{}{}
		}
	}

	// Check for dead service account usage
	for sa, b := range refs {
		ns, _ := namespaced(sa)
		if ns != "" && s.client.ExcludedNS(ns) {
			continue
		}
		s.initIssues(sa)
		if _, ok := psas[sa]; !ok {
			ns, n := namespaced(b)
			if ns == "" {
				s.addIssuef(sa, ErrorLevel, "Used? Referenced by CRB `%s", n)
			} else {
				s.addIssuef(sa, ErrorLevel, "Used? Referenced by RB `%s", n)
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Refs...

func namespaced(s string) (string, string) {
	tokens := strings.Split(s, "/")
	if len(tokens) == 2 {
		return tokens[0], tokens[1]
	}
	return "", tokens[0]
}

func pullSas(n string, ss []rbacv1.Subject, res map[string]string) {
	for _, s := range ss {
		if s.Kind == "ServiceAccount" {
			fqn := fqnSubjet(s)
			if _, ok := res[fqn]; !ok {
				res[fqn] = n
			}
		}
	}
}

func fqnSubjet(s rbacv1.Subject) string {
	if s.Namespace == "" {
		return s.Name
	}
	return s.Namespace + "/" + s.Name
}
