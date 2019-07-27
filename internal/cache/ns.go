package cache

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// ListNamespacesBySelector list all pods matching the given selector.
func (n *Namespace) ListNamespacesBySelector(sel *metav1.LabelSelector) map[string]*v1.Namespace {
	res := map[string]*v1.Namespace{}
	if sel == nil {
		return res
	}
	for fqn, ns := range n.nss {
		if matchLabels(ns.ObjectMeta.Labels, sel.MatchLabels) {
			res[fqn] = ns
		}
	}

	return res
}
