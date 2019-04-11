package linter

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

// SkipServices skips internal services for being included in scan.
// BOZO!! spinachyaml default??
var skipServices = []string{"default/kubernetes"}

// Service represents a service linter.
type Service struct {
	*Linter
}

// NewService returns a new service linter.
func NewService(l Loader, log *zerolog.Logger) *Service {
	return &Service{NewLinter(l, log)}
}

// Lint a service.
func (s *Service) Lint(ctx context.Context) error {
	services, err := s.ListServices()
	if err != nil {
		return err
	}

	for fqn, svc := range services {
		// Skip internal services...
		if in(skipServices, fqn) {
			continue
		}

		s.initIssues(fqn)
		po, err := s.GetPod(svc.Spec.Selector)
		if err != nil {
			s.addError(fqn, err)
		}
		ep, err := s.GetEndpoints(fqn)
		if err != nil {
			s.addError(fqn, err)
		}

		s.lint(svc, po, ep)
	}

	return nil
}

func (s *Service) lint(svc v1.Service, po *v1.Pod, ep *v1.Endpoints) {
	s.checkPorts(svc, po)
	s.checkEndpoints(svc, ep)
	s.checkType(svc)
}

func (s *Service) checkType(svc v1.Service) {
	if svc.Spec.Type == v1.ServiceTypeLoadBalancer {
		s.addIssue(svcFQN(svc), InfoLevel, "Type Loadbalancer detected. Could be expensive")
	}
}

func (s *Service) checkPorts(svc v1.Service, po *v1.Pod) {
	// No matching pod bail out!
	if po == nil && len(svc.Spec.Selector) > 0 {
		s.addError(svcFQN(svc), errors.New("No pods matched service selector"))
		return
	}

	for _, p := range svc.Spec.Ports {
		errs := checkServicePort(svc.Name, po, p)
		if errs != nil {
			s.addErrors(svcFQN(svc), errs...)
			continue
		}
	}
}

// CheckEndpoints runs a sanity check on this service endpoints.
func (s *Service) checkEndpoints(svc v1.Service, ep *v1.Endpoints) {
	if len(svc.Spec.Selector) == 0 {
		return
	}

	if svc.Spec.Type == v1.ServiceTypeExternalName {
		return
	}

	if ep == nil || len(ep.Subsets) == 0 {
		s.addError(svcFQN(svc), fmt.Errorf("No associated endpoints"))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func checkServicePort(svc string, pod *v1.Pod, port v1.ServicePort) []error {
	sPort := port.TargetPort.IntVal
	if sPort == 0 {
		sPort = port.Port
	}

	var errs []error
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if p.ContainerPort != sPort {
				continue
			}
			if p.Protocol != port.Protocol {
				errs = append(errs, fmt.Errorf("Port `%d protocol mismatch %s vs %s", sPort, port.Protocol, p.Protocol))
			}
			return nil
		}
	}

	return errs
}

func svcFQN(s v1.Service) string {
	return s.Namespace + "/" + s.Name
}
