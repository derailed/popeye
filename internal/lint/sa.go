// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const defaultSA = "default"

// ServiceAccount tracks ServiceAccount linter.
type ServiceAccount struct {
	*issues.Collector
	db *db.DB
}

// NewServiceAccount returns a new instance.
func NewServiceAccount(co *issues.Collector, db *db.DB) *ServiceAccount {
	return &ServiceAccount{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *ServiceAccount) Lint(ctx context.Context) error {
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

	txn, it := s.db.MustITFor(internal.Glossary[internal.SA])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		sa := o.(*v1.ServiceAccount)
		fqn := client.FQN(sa.Namespace, sa.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, sa))

		s.checkMounts(ctx, sa.AutomountServiceAccountToken)
		s.checkSecretRefs(ctx, fqn, sa.Secrets)
		s.checkPullSecretRefs(ctx, fqn, sa.ImagePullSecrets)
		if _, ok := refs[fqn]; !ok && sa.Name != defaultSA {
			s.AddCode(ctx, 400)
		}
	}

	return nil
}

func (s *ServiceAccount) checkSecretRefs(ctx context.Context, fqn string, refs []v1.ObjectReference) {
	ns, _ := namespaced(fqn)
	for _, ref := range refs {
		if ref.Namespace != "" {
			ns = ref.Namespace
		}
		sfqn := cache.FQN(ns, ref.Name)
		if !s.db.Exists(internal.Glossary[internal.SEC], sfqn) {
			s.AddCode(ctx, 304, sfqn)
		}
	}
}

func (s *ServiceAccount) checkPullSecretRefs(ctx context.Context, fqn string, refs []v1.LocalObjectReference) {
	ns, _ := namespaced(fqn)
	for _, ref := range refs {
		sfqn := cache.FQN(ns, ref.Name)
		if !s.db.Exists(internal.Glossary[internal.SEC], sfqn) {
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
	txn := s.db.Txn(false)
	defer txn.Abort()
	it, err := txn.Get(internal.Glossary[internal.CRB].String(), "id")
	if err != nil {
		return err
	}
	for o := it.Next(); o != nil; o = it.Next() {
		crb := o.(*rbacv1.ClusterRoleBinding)
		pullSas(crb.Subjects, refs)
	}

	return nil
}

func (s *ServiceAccount) rbRefs(refs map[string]struct{}) error {
	txn := s.db.Txn(false)
	defer txn.Abort()
	it, err := txn.Get(internal.Glossary[internal.ROB].String(), "id")
	if err != nil {
		return err
	}
	for o := it.Next(); o != nil; o = it.Next() {
		rb := o.(*rbacv1.RoleBinding)
		pullSas(rb.Subjects, refs)
	}

	return nil
}

func (s *ServiceAccount) podRefs(refs map[string]struct{}) error {
	txn := s.db.Txn(false)
	defer txn.Abort()
	it, err := txn.Get(internal.Glossary[internal.PO].String(), "id")
	if err != nil {
		return err
	}
	for o := it.Next(); o != nil; o = it.Next() {
		p := o.(*v1.Pod)
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
