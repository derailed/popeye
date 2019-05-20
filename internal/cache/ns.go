package cache

import (
	v1 "k8s.io/api/core/v1"
)

// Namespace represents a collection of Namespaces available on a cluster.
type Namespace struct {
	nss map[string]*v1.Namespace
}

// NewNamespace returns a new Namespace.
func NewNamespace(nss map[string]*v1.Namespace) *Namespace {
	return &Namespace{nss}
}

// ListNamespaces returns all available Namespaces on the cluster.
func (n *Namespace) ListNamespaces() map[string]*v1.Namespace {
	return n.nss
}
