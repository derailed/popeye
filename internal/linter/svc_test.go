package linter

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestSvcLint(t *testing.T) {
	uu := []struct {
		name   string
		port   v1.ServicePort
		ports  []int
		issues int
	}{
		{
			"fred",
			v1.ServicePort{Name: "s1", Port: 80, TargetPort: intstr.FromInt(80), Protocol: "TCP"},
			[]int{80, 90},
			0,
		},
		{
			"fred",
			v1.ServicePort{Name: "s1", Port: 80, TargetPort: intstr.FromInt(81), Protocol: "TCP"},
			[]int{80, 90},
			1,
		},
		{
			"fred",
			v1.ServicePort{Name: "s1", Port: 80, Protocol: "TCP"},
			[]int{80, 90},
			0,
		},
		{
			"fred",
			v1.ServicePort{Name: "s1", Port: 81, Protocol: "TCP"},
			[]int{80, 90},
			1,
		},
	}

	for _, u := range uu {
		svc := makeSvc("s1")
		svc.Spec.Ports = append(svc.Spec.Ports, u.port)
		po := makePod("fred")
		po.Spec.Containers = []v1.Container{
			{
				Name:  "c1",
				Ports: makePorts(u.ports...),
			},
		}
		s := NewService(nil, nil)
		ep := makeEp(svc.Name, "1.1.1.1")
		s.lint(svc, &po, &ep)

		assert.Equal(t, 0, len(s.Issues()[svcFQN(svc)]))
	}
}

func TestSvcCheckServicePort(t *testing.T) {
	uu := []struct {
		name   string
		port   v1.ServicePort
		ports  []int
		issues int
	}{
		{
			"fred",
			v1.ServicePort{Name: "s1", Port: 80, TargetPort: intstr.FromInt(80), Protocol: "TCP"},
			[]int{80, 90},
			0,
		},
		{
			"fred",
			v1.ServicePort{Name: "s1", Port: 80, TargetPort: intstr.FromInt(81), Protocol: "TCP"},
			[]int{80, 90},
			1,
		},
		{
			"fred",
			v1.ServicePort{Name: "s1", Port: 80, Protocol: "TCP"},
			[]int{80, 90},
			0,
		},
		{
			"fred",
			v1.ServicePort{Name: "s1", Port: 81, Protocol: "TCP"},
			[]int{80, 90},
			1,
		},
	}

	for _, u := range uu {
		svc := makeSvc("s1")
		svc.Spec.Ports = append(svc.Spec.Ports, u.port)
		po := makePod("fred")
		po.Spec.Containers = []v1.Container{
			{
				Name:  "c1",
				Ports: makePorts(u.ports...),
			},
		}
		s := NewService(nil, nil)
		s.checkPorts(svc, &po)

		assert.Equal(t, 0, len(s.Issues()[svcFQN(svc)]))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeSvc(s string) v1.Service {
	return v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s,
			Namespace: "default",
		},
	}
}
func makePorts(ports ...int) []v1.ContainerPort {
	pp := []v1.ContainerPort{}
	for _, p := range ports {
		pp = append(pp, v1.ContainerPort{
			Name:          "p" + strconv.Itoa(int(p)),
			ContainerPort: int32(p),
			Protocol:      v1.ProtocolTCP,
		})
	}
	return pp
}

func makeEp(s string, ips ...string) v1.Endpoints {
	ep := v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s,
			Namespace: "default",
		},
	}

	var add []v1.EndpointAddress
	for _, ip := range ips {
		add = append(add, v1.EndpointAddress{IP: ip})
	}
	ep.Subsets = []v1.EndpointSubset{
		{Addresses: add},
	}
	return ep
}
