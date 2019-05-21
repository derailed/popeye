package dag

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// MetaFQN returns a full qualified ns/name string.
func metaFQN(m metav1.ObjectMeta) string {
	if m.Namespace == "" {
		return m.Name
	}

	return m.Namespace + "/" + m.Name
}
