package generated

import (
	"github.com/derailed/popeye/internal/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Namespace represents a Kubernetes Namespace.
type Namespace struct{}

// List all Namespaces.
func (*Namespace) List(c *k8s.Client) (*v1.NamespaceList, error) {
	var list v1.NamespaceList
	dial, err := c.Dial()
	if err != nil {
		return &list, err
	}

	return dial.CoreV1().Namespaces().List(metav1.ListOptions{})
}

// Get a Namespace.
func (*Namespace) Get(c *k8s.Client, name string) (*v1.Namespace, error) {
	var res v1.Namespace
	dial, err := c.Dial()
	if err != nil {
		return &res, err
	}

	return dial.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
}