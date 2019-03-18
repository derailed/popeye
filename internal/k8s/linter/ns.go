package linter

import (
	v1 "k8s.io/api/core/v1"
)

// Namespace represents a Namespace linter.
type Namespace struct {
	*Linter
}

// NewNamespace returns a new namespace linter.
func NewNamespace() *Namespace {
	return &Namespace{new(Linter)}
}

// Lint a namespace
func (n *Namespace) Lint(ns v1.Namespace) {
	n.checkActive(ns)
}

func (n *Namespace) checkActive(ns v1.Namespace) {
	if ns.Status.Phase != v1.NamespaceActive {
		n.addIssuef(ErrorLevel, "namespace %s is inactive", ns.Name)
	}
}
