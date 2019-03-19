package generated

import (
	"github.com/derailed/popeye/internal/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// {{ .Resource }} represents a Kubernetes {{ .Resource }}.
type {{ .Resource }} struct{}

// List all {{ .Resource }}s.
func (*{{ .Resource }}) List(c *k8s.Client) (*{{ .Version }}.{{ .Resource }}List, error) {
	var list {{ .Version }}.{{ .Resource }}List
	dial, err := c.Dial()
	if err != nil {
		return &list, err
	}

	return dial.CoreV1().{{ .Resource }}s().List(metav1.ListOptions{})
}

// Get a {{ .Resource }}.
func (*{{ .Resource }}) Get(c *k8s.Client, name string) (*{{ .Version }}.{{ .Resource }}, error) {
	var res {{ .Version }}.{{ .Resource }}
	dial, err := c.Dial()
	if err != nil {
		return &res, err
	}

	return dial.CoreV1().{{ .Resource }}s().Get(name, metav1.GetOptions{})
}