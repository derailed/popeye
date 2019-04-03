package linter

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var skipServices = []string{"default/kubernetes"}

// Check port mappings
// Check endpoints
// Check LoadBalancer type

// Service represents a service linter.
type Service struct {
	*Linter
}

// NewService returns a new service linter.
func NewService(c *k8s.Client, l *zerolog.Logger) *Service {
	return &Service{newLinter(c, l)}
}

// Lint a service.
func (s *Service) Lint(ctx context.Context) error {
	list, err := s.client.DialOrDie().
		CoreV1().
		Services(s.client.Config.ActiveNamespace()).
		List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, svc := range list.Items {
		// Skip internal services...
		if in(skipServices, svcFQN(svc)) {
			continue
		}

		s.initIssues(svcFQN(svc))

		po, err := s.findPod(svc)
		if err != nil {
			s.addError(svcFQN(svc), err)
		}

		ep, err := s.client.DialOrDie().
			CoreV1().
			Endpoints(svc.Namespace).
			Get(svc.Name, metav1.GetOptions{})
		if err != nil {
			s.addError(svcFQN(svc), err)
		}
		s.lint(svc, po, ep)
	}

	return nil
}

func (s *Service) lint(svc v1.Service, po *v1.Pod, ep *v1.Endpoints) {
	if po != nil {
		s.checkPorts(svc, po)
	}
	if ep != nil {
		s.checkEndpoints(svc, ep)
	}
}

func (s *Service) checkPorts(svc v1.Service, po *v1.Pod) {
	for _, p := range svc.Spec.Ports {
		errs := checkServicePort(svc.Name, po, p)
		if errs != nil {
			s.addErrors(svcFQN(svc), errs)
			continue
		}
	}
}

// CheckEndpoints runs a sanity check on all endpoints in a given namespace.
func (s *Service) checkEndpoints(svc v1.Service, ep *v1.Endpoints) {
	if svc.Spec.Type == v1.ClusterIPNone {
		return
	}

	if len(ep.Subsets) == 0 {
		s.addError(svcFQN(svc), fmt.Errorf("No associated endpoints"))
	}
}

func (s *Service) findPod(svc v1.Service) (*v1.Pod, error) {
	pods, err := s.client.DialOrDie().CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: toSelector(svc.Spec.Selector),
	})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("No pods match service selector")
	}

	return &pods.Items[0], nil
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
			if p.ContainerPort == sPort {
				if p.Protocol != port.Protocol {
					errs = append(errs, fmt.Errorf("Port `%d protocol mismatch %s vs %s", sPort, port.Protocol, p.Protocol))
				}
				return nil
			}
		}
	}

	return errs
}

func svcFQN(s v1.Service) string {
	return s.Namespace + "/" + s.Name
}

func toSelector(labels map[string]string) string {
	// Ensure no matches!
	if len(labels) == 0 {
		return "bozo=xxx"
	}

	ss := make([]string, 0, len(labels))
	for k, v := range labels {
		ss = append(ss, k+"="+v)
	}
	return strings.Join(ss, ",")
}

// In checks if an item is in a list of items.
func in(ll []string, s string) bool {
	for _, l := range ll {
		if l == s {
			return true
		}
	}

	return false
}
