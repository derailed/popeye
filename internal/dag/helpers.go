package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetaFQN returns a full qualified ns/name string.
func metaFQN(m metav1.ObjectMeta) string {
	if m.Namespace == "" {
		return m.Name
	}

	return m.Namespace + "/" + m.Name
}

// IncludeNS checks if namespace should be included.
func includeNS(c *k8s.Client, ns string) bool {
	return c.IsActiveNamespace(ns)
}
