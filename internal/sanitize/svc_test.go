package sanitize

import (
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestSVCSanitize(t *testing.T) {
	uu := map[string]struct {
		lister ServiceLister
		issues int
	}{
		"cool": {
			makeSvcLister(
				svcOpts{
					kind:         v1.ServiceTypeClusterIP,
					hasEndPoints: true,
					hasSelector:  true,
					hasPod:       true,
				},
			),
			0,
		},
		"noEp": {
			makeSvcLister(
				svcOpts{
					kind:        v1.ServiceTypeClusterIP,
					hasSelector: true,
					hasPod:      true,
				},
			),
			1,
		},
		"noMatchingPods": {
			makeSvcLister(
				svcOpts{
					kind:         v1.ServiceTypeClusterIP,
					hasSelector:  true,
					hasEndPoints: true,
				},
			),
			1,
		},
		"lbType": {
			makeSvcLister(
				svcOpts{
					kind:         v1.ServiceTypeLoadBalancer,
					hasEndPoints: true,
					hasSelector:  true,
					hasPod:       true,
				},
			),
			1,
		},
		"nodePortType": {
			makeSvcLister(
				svcOpts{
					kind:         v1.ServiceTypeNodePort,
					hasEndPoints: true,
					hasSelector:  true,
					hasPod:       true,
				},
			),
			1,
		},
		"noSelector": {
			makeSvcLister(
				svcOpts{
					kind:         v1.ServiceTypeClusterIP,
					hasEndPoints: true,
					hasPod:       true,
				},
			),
			0,
		},
		"externalSvc": {
			makeSvcLister(
				svcOpts{
					kind:        v1.ServiceTypeExternalName,
					hasSelector: true,
					hasPod:      true,
				},
			),
			0,
		},
		"portProtoFail": {
			makeSvcLister(
				svcOpts{
					kind:        v1.ServiceTypeExternalName,
					hasSelector: true,
					hasPod:      true,
					ports: []v1.ServicePort{
						{
							Name:       "p1",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolUDP,
						},
					},
				},
			),
			1,
		},
		"badTargetPortNumb": {
			makeSvcLister(
				svcOpts{
					kind:         v1.ServiceTypeClusterIP,
					hasSelector:  true,
					hasPod:       true,
					hasEndPoints: true,
					ports: []v1.ServicePort{
						{
							Name:       "p1",
							Port:       80,
							TargetPort: intstr.Parse("90"),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			),
			1,
		},
		"badNamedTargetPort": {
			makeSvcLister(
				svcOpts{
					kind:         v1.ServiceTypeClusterIP,
					hasSelector:  true,
					hasPod:       true,
					hasEndPoints: true,
					ports: []v1.ServicePort{
						{
							Name:       "p1",
							Port:       80,
							TargetPort: intstr.Parse("toast"),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			),
			1,
		},
		"unnamedTargetPort": {
			makeSvcLister(
				svcOpts{
					kind:         v1.ServiceTypeClusterIP,
					hasSelector:  true,
					hasPod:       true,
					hasEndPoints: true,
					ports: []v1.ServicePort{
						{
							Name:       "p1",
							Port:       80,
							TargetPort: intstr.Parse("80"),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			),
			1,
		},
		"unamedSvcPort": {
			makeSvcLister(
				svcOpts{
					kind:         v1.ServiceTypeClusterIP,
					hasSelector:  true,
					hasPod:       true,
					hasEndPoints: true,
					ports: []v1.ServicePort{
						{
							Port:       80,
							Protocol:   v1.ProtocolTCP,
							TargetPort: intstr.Parse("p1"),
						},
					},
				},
			),
			0,
		},
		"unmatchedSvcPort": {
			makeSvcLister(
				svcOpts{
					kind:         v1.ServiceTypeClusterIP,
					hasSelector:  true,
					hasPod:       true,
					hasEndPoints: true,
					ports: []v1.ServicePort{
						{
							Name:       "p3",
							Port:       15014,
							Protocol:   v1.ProtocolTCP,
							TargetPort: intstr.Parse("15014"),
						},
					},
				},
			),
			1,
		},
	}

	ctx := makeContext("v1/services", "svc")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := NewService(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, s.Sanitize(ctx))
			assert.Equal(t, u.issues, len(s.Outcome()["default/s1"]))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type (
	svcOpts struct {
		hasEndPoints bool
		hasPod       bool
		hasSelector  bool
		kind         v1.ServiceType
		ports        []v1.ServicePort
	}

	svc struct {
		name string
		opts svcOpts
	}
)

func makeSvcLister(opts svcOpts) *svc {
	return &svc{
		name: "s1",
		opts: opts,
	}
}

func (s *svc) ListServices() map[string]*v1.Service {
	return map[string]*v1.Service{
		cache.FQN("default", s.name): makeSvc(s.name, s.opts),
	}
}

func (s *svc) GetPod(string, map[string]string) *v1.Pod {
	if s.opts.hasPod {
		return makeSvcPod("p1")
	}

	return nil
}

func (s *svc) GetEndpoints(string) *v1.Endpoints {
	if s.opts.hasEndPoints {
		return makeEp(s.name, []string{"1.1.1.1", "2.2.2.2"}...)
	}

	return nil
}

func makeSvcPod(n string) *v1.Pod {
	po := makePod(n)

	po.Spec = v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:  "c1",
				Image: "freddy:0.0.1",
				Ports: []v1.ContainerPort{
					{Name: "p1", ContainerPort: 80, Protocol: v1.ProtocolTCP},
					{Name: "p2", ContainerPort: 81, Protocol: v1.ProtocolUDP},
				},
			},
		},
		InitContainers: []v1.Container{
			{
				Name:  "i1",
				Image: "freddo:0.0.1",
			},
		},
	}

	return po
}

func makeSvc(s string, opts svcOpts) *v1.Service {
	svc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s,
			Namespace: "default",
		},
		Spec: v1.ServiceSpec{
			Type: opts.kind,
		},
	}
	if opts.hasSelector {
		svc.Spec.Selector = map[string]string{"app": "fred"}
	}
	svc.Spec.Ports = opts.ports

	return &svc
}

func makeEp(s string, ips ...string) *v1.Endpoints {
	ep := &v1.Endpoints{
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
