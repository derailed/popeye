// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"strconv"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Service represents a service linter.
type Service struct {
	*issues.Collector
	db *db.DB
}

// NewService returns a new instance.
func NewService(co *issues.Collector, db *db.DB) *Service {
	return &Service{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *Service) Lint(ctx context.Context) error {
	txn, it := s.db.MustITFor(internal.Glossary[internal.SVC])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		svc := o.(*v1.Service)
		fqn := client.FQN(svc.Namespace, svc.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, svc))

		if len(svc.Spec.Selector) > 0 {
			s.checkPorts(ctx, svc.Namespace, svc.Spec.Selector, svc.Spec.Ports)
			s.checkEndpoints(ctx, fqn, svc.Spec.Type)
		}
		s.checkType(ctx, svc.Spec.Type)
		s.checkExternalTrafficPolicy(ctx, svc.Spec.Type, svc.Spec.ExternalTrafficPolicy)
	}

	return nil
}

func (s *Service) checkPorts(ctx context.Context, ns string, sel map[string]string, ports []v1.ServicePort) {
	po, err := s.db.FindPod(ns, sel)
	if err != nil || po == nil {
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
func (s *Service) checkEndpoints(ctx context.Context, fqn string, kind v1.ServiceType) {
	// External service bail -> no EPs.
	if kind == v1.ServiceTypeExternalName {
		return
	}

	o, err := s.db.Find(internal.Glossary[internal.EP], fqn)
	if err != nil {
		s.AddCode(ctx, 1105)
		return
	}
	ep := o.(*v1.Endpoints)
	if len(ep.Subsets) == 0 {
		s.AddCode(ctx, 1110)
		return
	}
	var eps int
	for _, s := range ep.Subsets {
		eps += len(s.Addresses)
	}
	if eps == 1 {
		s.AddCode(ctx, 1109)
	}
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
	cos := append([]v1.Container{}, pod.Spec.Containers...)

	for _, ico := range pod.Spec.InitContainers {
		if ico.RestartPolicy != nil && *ico.RestartPolicy == v1.ContainerRestartPolicyAlways {
			cos = append(cos, ico)
		}
	}

	for _, co := range cos {
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
