package sanitize

import (
	"context"
	"errors"
	"fmt"

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
			s.AddErr(fqn, errors.New("No pods matched service selector"))
		}
		return
	}

	for _, p := range ports {
		errs := checkServicePort(p, po)
		if errs != nil {
			s.AddErr(fqn, errs...)
			continue
		}
	}
}

func (s *Service) checkType(fqn string, kind v1.ServiceType) {
	if kind == v1.ServiceTypeLoadBalancer {
		s.AddInfo(fqn, "Type Loadbalancer detected. Could be expensive")
	}
	if kind == v1.ServiceTypeNodePort {
		s.AddInfo(fqn, "Type NodePort detected. Do mean it?")
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
		s.AddErr(fqn, errors.New("No associated endpoints"))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

// CheckServicePort
func checkServicePort(port v1.ServicePort, pod *v1.Pod) []error {
	var errs []error
	var match bool
	for _, co := range pod.Spec.Containers {
		for _, p := range co.Ports {
			if !matchPort(port, p) {
				continue
			}
			match = true
			if p.Protocol != port.Protocol {
				errs = append(
					errs,
					fmt.Errorf("Port `%s protocol mismatch %s vs %s", portAsStr(port), port.Protocol, p.Protocol),
				)
			}
		}
	}

	if !match {
		errs = append(
			errs,
			fmt.Errorf("No container ports matches service port `%s", portAsStr(port)),
		)
	}

	return errs
}

// MatchPort check if service port matches a given container port.
// Return true if service port or target port matches container port, false otherwise.
func matchPort(sp v1.ServicePort, cp v1.ContainerPort) bool {
	found, hasTarget := false, false
	switch sp.TargetPort.Type {
	case intstr.Int:
		if sp.TargetPort.IntValue() != 0 {
			hasTarget = true
			found = sp.TargetPort.IntValue() == int(cp.ContainerPort)
		}
	case intstr.String:
		if sp.TargetPort.String() != "" {
			hasTarget = true
			found = sp.TargetPort.String() == cp.Name
		}
	}
	if found {
		return found
	}
	if hasTarget && !found {
		return found
	}
	if sp.Port == cp.ContainerPort {
		return true
	}

	return false
}
