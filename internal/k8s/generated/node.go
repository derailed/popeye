package generated

import (
	"github.com/derailed/popeye/internal/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Node represents a Kubernetes Node.
type Node struct{}

// List all Nodes.
func (*Node) List(c *k8s.Client) (*v1.NodeList, error) {
	var list v1.NodeList
	dial, err := c.Dial()
	if err != nil {
		return &list, err
	}

	return dial.CoreV1().Nodes().List(metav1.ListOptions{})
}

// Get a Node.
func (*Node) Get(c *k8s.Client, name string) (*v1.Node, error) {
	var res v1.Node
	dial, err := c.Dial()
	if err != nil {
		return &res, err
	}

	return dial.CoreV1().Nodes().Get(name, metav1.GetOptions{})
}