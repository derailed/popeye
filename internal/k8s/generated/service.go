package generated

import (
	"github.com/derailed/popeye/internal/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Service represents a Kubernetes Service.
type Service struct{}

// List all Services.
func (*Service) List(conn *k8s.Server, ns string) (*v1.ServiceList, error) {
	var list v1.ServiceList
	dial, err := conn.Dial()
	if err != nil {
		return &list, err
	}

	return dial.CoreV1().Services(ns).List(metav1.ListOptions{})
}

// Get a Service.
func (*Service) Get(conn *k8s.Server, name, ns string) (*v1.Service, error) {
	var res v1.Service
	dial, err := conn.Dial()
	if err != nil {
		return &res, err
	}

	return dial.CoreV1().Services(ns).Get(name, metav1.GetOptions{})
}