package sanitize

import (
	"context"
	"fmt"
	"strconv"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// SkipServices skips internal services for being included in scan.
// BOZO!! spinachyaml default??
var skipServices = []string{"default/kubernetes"}

type (
	// ServiceLister list available Services on a cluster.
	ServiceLister interface {
		PodGetter
		EndPointLister
		ListServices() map[string]*v1.Service
	}

	// PodGetter find a single pod matching service selector.
	PodGetter interface {
		GetPod(map[string]string) *v1.Pod
	}

	// EndPointLister find all service endpoints.
	EndPointLister interface {
		GetEndpoints(string) *v1.Endpoints
	}

	// Service represents a service linter.
	Service struct {
		*issues.Collector
		ServiceLister
	}
)

// NewService returns a new service linter.
func NewService(co *issues.Collector, lister ServiceLister) *Service {
	return &Service{
		Collector:     co,
		ServiceLister: lister,
	}
}

// Sanitize services.
func (s *Service) Sanitize(ctx context.Context) error {
	for fqn, svc := range s.ListServices() {
		// Skip internal services...
		if in(skipServices, fqn) {
			continue
		}
		s.InitOutcome(fqn)
		s.checkPorts(fqn, svc.Spec.Selector, svc.Spec.Ports)
		s.checkEndpoints(fqn, svc.Spec.Selector, svc.Spec.Type)
		s.checkType(fqn, svc.Spec.Type)
	}

	return nil
}

func (s *Service) checkPorts(fqn string, sel map[string]string, ports []v1.ServicePort) {
	po := s.GetPod(sel)
	if po == nil {
		if len(sel) > 0 {
			s.AddCode(1100, fqn)
		}
		return
	}

	pports := make(map[string]string)
	portsForPod(po, pports)
	pfqn := cache.MetaFQN(po.ObjectMeta)
	// No explicit pod ports definition -> bail!.
	if len(pports) == 0 {
		s.AddCode(1101, fqn, pfqn)
		return
	}
	for _, p := range ports {
		err := checkServicePort(p, pports)
		if err != nil {
			s.AddErr(fqn, err)
			continue
		}
		if !checkNamedTargetPort(p) {
			s.AddCode(1102, fqn, p.TargetPort.String(), portAsStr(p))
		}
	}
}

func (s *Service) checkType(fqn string, kind v1.ServiceType) {
	if kind == v1.ServiceTypeLoadBalancer {
		s.AddCode(1103, fqn)
	}
	if kind == v1.ServiceTypeNodePort {
		s.AddCode(1104, fqn)
	}
}

// CheckEndpoints runs a sanity check on this service endpoints.
func (s *Service) checkEndpoints(fqn string, sel map[string]string, kind v1.ServiceType) {
	// Service may not have selectors.
	if len(sel) == 0 {
		return
	}
	// External service bail -> no EPs.
	if kind == v1.ServiceTypeExternalName {
		return
	}
	ep := s.GetEndpoints(fqn)
	if ep == nil || len(ep.Subsets) == 0 {
		s.AddCode(1105, fqn)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func checkNamedTargetPort(port v1.ServicePort) bool {
	return port.TargetPort.Type == intstr.String
}

// CheckServicePort
func checkServicePort(port v1.ServicePort, ports map[string]string) error {
	fqn := servicePortFQN(port)
	if _, ok := ports[fqn]; ok {
		return nil
	}

	return fmt.Errorf("No target ports match service port `%s", portAsStr(port))
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
