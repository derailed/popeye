package generated

import (
	"github.com/derailed/popeye/internal/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Namespace represents a Kubernetes Namespace.
type Namespace struct{}

// List all Namespaces.
func (*Namespace) List(conn *k8s.Server) (*v1.NamespaceList, error) {
	var list v1.NamespaceList
	dial, err := conn.Dial()
	if err != nil {
		return &list, err
	}

	return dial.CoreV1().Namespaces().List(metav1.ListOptions{})
}

// Get a Namespace.
func (*Namespace) Get(conn *k8s.Server, name string) (*v1.Namespace, error) {
	var res v1.Namespace
	dial, err := conn.Dial()
	if err != nil {
		return &res, err
	}

	return dial.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
}