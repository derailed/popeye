package sanitize

import (
	"context"
	"strconv"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type (
	// ServiceLister list available Services on a cluster.
	ServiceLister interface {
		PodGetter
		EndPointLister
		ListServices() map[string]*v1.Service
	}

	// PodGetter find a single pod matching service selector.
	PodGetter interface {
		GetPod(ns string, sel map[string]string) *v1.Pod
	}

	// EndPointLister find all service endpoints.
	EndPointLister interface {
		GetEndpoints(string) *v1.Endpoints
	}

	// Service represents a service sanitizer.
	Service struct {
		*issues.Collector
		ServiceLister
	}
)

// NewService returns a new sanitizer.
func NewService(co *issues.Collector, lister ServiceLister) *Service {
	return &Service{
		Collector:     co,
		ServiceLister: lister,
	}
}

// Sanitize cleanse the resource.
func (s *Service) Sanitize(ctx context.Context) error {
	for fqn, svc := range s.ListServices() {
		s.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		s.checkPorts(ctx, svc.Namespace, svc.Spec.Selector, svc.Spec.Ports)
		s.checkEndpoints(ctx, svc.Spec.Selector, svc.Spec.Type)
		s.checkType(ctx, svc.Spec.Type)
		s.checkExternalTrafficPolicy(ctx, svc.Spec.Type, svc.Spec.ExternalTrafficPolicy)

		if s.NoConcerns(fqn) && s.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			s.ClearOutcome(fqn)
		}
	}

	return nil
}

func (s *Service) checkPorts(ctx context.Context, ns string, sel map[string]string, ports []v1.ServicePort) {
	po := s.GetPod(ns, sel)
	if po == nil {
		if len(sel) > 0 {
			s.AddCode(ctx, 1100)
		}
		return
	}

	pports := make(map[string]string)
	portsForPod(po, pports)
	pfqn := cache.MetaFQN(po.ObjectMeta)
	// No explicit pod ports definition -> bail!.
	if len(pports) == 0 {
		s.AddCode(ctx, 1101, pfqn)
		return
	}
	for _, p := range ports {
		if !checkServicePort(p, pports) {
			s.AddCode(ctx, 1106, portAsStr(p))
			continue
		}
		if !checkNamedTargetPort(p) {
			s.AddCode(ctx, 1102, p.TargetPort.String(), portAsStr(p))
		}
	}
}

func (s *Service) checkType(ctx context.Context, kind v1.ServiceType) {
	if kind == v1.ServiceTypeLoadBalancer {
		s.AddCode(ctx, 1103)
	}
	if kind == v1.ServiceTypeNodePort {
		s.AddCode(ctx, 1104)
	}
}

func (s *Service) checkExternalTrafficPolicy(ctx context.Context, kind v1.ServiceType, policy v1.ServiceExternalTrafficPolicyType) {
	if kind == v1.ServiceTypeLoadBalancer && policy == v1.ServiceExternalTrafficPolicyTypeCluster {
		s.AddCode(ctx, 1107)
		return
	}
	if kind == v1.ServiceTypeNodePort && policy == v1.ServiceExternalTrafficPolicyTypeLocal {
		s.AddCode(ctx, 1108)
	}
}

// CheckEndpoints runs a sanity check on this service endpoints.
func (s *Service) checkEndpoints(ctx context.Context, sel map[string]string, kind v1.ServiceType) {
	// Service may not have selectors.
	if len(sel) == 0 {
		return
	}
	// External service bail -> no EPs.
	if kind == v1.ServiceTypeExternalName {
		return
	}
	ep := s.GetEndpoints(internal.MustExtractFQN(ctx))
	if ep == nil || len(ep.Subsets) == 0 {
		s.AddCode(ctx, 1105)
		return
	}
	numEndpointAddresses := 0
	for _, s := range ep.Subsets {
		numEndpointAddresses += len(s.Addresses)
		if numEndpointAddresses > 1 {
			return
		}
	}
	s.AddCode(ctx, 1109)
}

// ----------------------------------------------------------------------------
// Helpers...

func checkNamedTargetPort(port v1.ServicePort) bool {
	return port.TargetPort.Type == intstr.String
}

// CheckServicePort
func checkServicePort(port v1.ServicePort, ports map[string]string) bool {
	fqn := servicePortFQN(port)
	if _, ok := ports[fqn]; ok {
		return true
	}

	return false
}

// PortsForPod computes a port map for a given pod.
func portsForPod(pod *v1.Pod, ports map[string]string) {
	for _, co := range pod.Spec.Containers {
		for _, p := range co.Ports {
			ports[portFQN(p.Protocol, strconv.Itoa(int(p.ContainerPort)))] = co.Name
			if p.Name != "" {
				ports[portFQN(p.Protocol, p.Name)] = co.Name
			}
		}
	}
}

func servicePortFQN(port v1.ServicePort) string {
	if port.TargetPort.String() != "" {
		return portFQN(port.Protocol, port.TargetPort.String())
	}

	return portFQN(port.Protocol, strconv.Itoa(int(port.Port)))
}

func portFQN(p v1.Protocol, port string) string {
	return string(p) + ":" + port
}
