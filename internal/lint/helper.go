// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/rules"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	qosBestEffort qos = iota
	qosBurstable
	qosGuaranteed

	// MegaByte represents a Mb.
	megaByte = 1024 * 1024
)

type qos = int

func coSpecFor(fqn string, o metav1.ObjectMetaAccessor, spec v1.PodSpec) rules.Spec {
	rule := SpecFor(fqn, o)
	rule.Containers = fetchContainers(spec)

	return rule
}

func fetchContainers(podTemplate v1.PodSpec) []string {
	containers := make([]string, 0, len(podTemplate.InitContainers)+len(podTemplate.Containers))
	for _, co := range podTemplate.InitContainers {
		containers = append(containers, co.Name)
	}
	for _, co := range podTemplate.Containers {
		containers = append(containers, co.Name)
	}

	return containers
}

// SpecFor construct a new run spec for a given resource.
func SpecFor(fqn string, o metav1.ObjectMetaAccessor) rules.Spec {
	spec := rules.Spec{
		FQN: fqn,
	}
	if o == nil {
		return spec
	}

	m := o.GetObjectMeta()
	spec.Labels, spec.Annotations = m.GetLabels(), m.GetAnnotations()

	return spec
}

func resourceUsage(ctx context.Context, dba *db.DB, c Collector, ns string, sel *metav1.LabelSelector) ConsumptionMetrics {
	var mx ConsumptionMetrics

	pp, err := dba.FindPodsBySel(ns, sel)
	if err != nil {
		c.AddErr(ctx, err)
		return mx
	}
	if len(pp) == 0 {
		c.AddCode(ctx, 508, dumpSel(sel))
		return mx
	}

	for _, pod := range pp {
		fqn := cache.FQN(pod.Namespace, pod.Name)
		cpu, mem := computePodResources(pod.Spec)
		mx.QOS = pod.Status.QOSClass
		mx.RequestCPU.Add(cpu)
		mx.RequestMEM.Add(mem)

		pmx, err := dba.FindPMX(fqn)
		if err != nil || pmx == nil {
			continue
		}
		for _, cx := range pmx.Containers {
			mx.CurrentCPU.Add(*cx.Usage.Cpu())
			mx.CurrentMEM.Add(*cx.Usage.Memory())
		}
	}

	return mx
}

// Poor man plural...
func pluralOf(s string, count int) string {
	if count > 1 {
		return s + "s"
	}
	return s
}

// Namespaced pull namespace and name out of an fqn.
func namespaced(s string) (string, string) {
	tokens := strings.Split(s, "/")
	if len(tokens) == 2 {
		return tokens[0], tokens[1]
	}

	return "", tokens[0]
}

// ToPerc computes the percentage from one number over another.
func ToPerc(v1, v2 int64) int64 {
	if v2 == 0 {
		return 0
	}
	return int64((float64(v1) / float64(v2)) * 100)
}

// ToMC converts quantity to millicores.
func toMC(q resource.Quantity) int64 {
	return q.MilliValue()
}

// ToMB converts quantity to megabytes.
func toMB(q resource.Quantity) int64 {
	return q.Value() / megaByte
}

// AsPec prints value as percentage.
func asPerc(n float64) string {
	return fmt.Sprintf("%0.2f%%", n)
}

// ToMCRatio computes millicore ratio.
func toMCRatio(q1, q2 resource.Quantity) float64 {
	if q2.IsZero() {
		return 0
	}
	v1, v2 := toMC(q1), toMC(q2)
	return (float64(v1) / float64(v2)) * 100
}

// ToMEMRatio computes mem MB ratio.
func toMEMRatio(q1, q2 resource.Quantity) float64 {
	if q2.IsZero() {
		return 0
	}
	v1, v2 := toMB(q1), toMB(q2)

	return (float64(v1) / float64(v2)) * 100
}

// AsMC prints millicore value.
func asMC(q resource.Quantity) string {
	return fmt.Sprintf("%vm", toMC(q))
}

// AsMB prints MB value.
func asMB(q resource.Quantity) string {
	return fmt.Sprintf("%vMi", toMB(q))
}

// PodResources computes pod resources as sum of containers allocations.
func podResources(spec v1.PodSpec) (cpu, mem resource.Quantity) {
	for _, co := range spec.InitContainers {
		c, m, _ := containerResources(co)
		if c != nil {
			cpu.Add(*c)
		}
		if m != nil {
			mem.Add(*m)
		}
	}

	for _, co := range spec.Containers {
		c, m, _ := containerResources(co)
		if c != nil {
			cpu.Add(*c)
		}
		if m != nil {
			mem.Add(*m)
		}
	}

	return
}

// ContainerResources gathers container resources setting.
func containerResources(co v1.Container) (cpu, mem *resource.Quantity, qos qos) {
	req, limit, qos := co.Resources.Requests, co.Resources.Limits, qosBurstable
	switch {
	case len(req) != 0 && len(limit) != 0:
		if req.Cpu().Cmp(*limit.Cpu()) == 0 && req.Memory().Cmp(*limit.Memory()) == 0 {
			qos = qosGuaranteed
		}
		cpu, mem = req.Cpu(), req.Memory()
	case len(req) != 0:
		cpu, mem = req.Cpu(), req.Memory()
	case len(limit) != 0:
		cpu, mem = limit.Cpu(), limit.Memory()
	default:
		qos = qosBestEffort
	}

	return
}

// PortAsString prints service port name or number.
func portAsStr(p v1.ServicePort) string {
	if p.Name != "" {
		return string(p.Protocol) + ":" + p.Name + ":" + strconv.Itoa(int(p.Port))
	}
	return string(p.Protocol) + "::" + strconv.Itoa(int(p.Port))
}

const (
	nodeUnreachablePodReason = "NodeLost"
	completed                = "Completed"
	running                  = "Running"
	terminating              = "Terminating"
)

func Phase(po *v1.Pod) string {
	status := string(po.Status.Phase)
	if po.Status.Reason != "" {
		if po.DeletionTimestamp != nil && po.Status.Reason == nodeUnreachablePodReason {
			return "Unknown"
		}
		status = po.Status.Reason
	}

	status, ok := initContainerPhase(po, status)
	if ok {
		return status
	}

	status, ok = containerPhase(po.Status, status)
	if ok && status == completed {
		status = running
	}
	if po.DeletionTimestamp == nil {
		return status
	}

	return terminating
}

func containerPhase(st v1.PodStatus, status string) (string, bool) {
	var running bool
	for i := len(st.ContainerStatuses) - 1; i >= 0; i-- {
		cs := st.ContainerStatuses[i]
		switch {
		case cs.State.Waiting != nil && cs.State.Waiting.Reason != "":
			status = cs.State.Waiting.Reason
		case cs.State.Terminated != nil && cs.State.Terminated.Reason != "":
			status = cs.State.Terminated.Reason
		case cs.State.Terminated != nil:
			if cs.State.Terminated.Signal != 0 {
				status = "Signal:" + strconv.Itoa(int(cs.State.Terminated.Signal))
			} else {
				status = "ExitCode:" + strconv.Itoa(int(cs.State.Terminated.ExitCode))
			}
		case cs.Ready && cs.State.Running != nil:
			running = true
		}
	}

	return status, running
}

func initContainerPhase(po *v1.Pod, status string) (string, bool) {
	count := len(po.Spec.InitContainers)
	rs := make(map[string]bool, count)
	for _, c := range po.Spec.InitContainers {
		rs[c.Name] = restartableInitCO(c.RestartPolicy)
	}
	for i, cs := range po.Status.InitContainerStatuses {
		if s := checkInitContainerStatus(cs, i, count, rs[cs.Name]); s != "" {
			return s, true
		}
	}

	return status, false
}

func restartableInitCO(p *v1.ContainerRestartPolicy) bool {
	return p != nil && *p == v1.ContainerRestartPolicyAlways
}

func checkInitContainerStatus(cs v1.ContainerStatus, count, initCount int, restartable bool) string {
	switch {
	case cs.State.Terminated != nil:
		if cs.State.Terminated.ExitCode == 0 {
			return ""
		}
		if cs.State.Terminated.Reason != "" {
			return "Init:" + cs.State.Terminated.Reason
		}
		if cs.State.Terminated.Signal != 0 {
			return "Init:Signal:" + strconv.Itoa(int(cs.State.Terminated.Signal))
		}
		return "Init:ExitCode:" + strconv.Itoa(int(cs.State.Terminated.ExitCode))
	case restartable && cs.Started != nil && *cs.Started:
		if cs.Ready {
			return ""
		}
	case cs.State.Waiting != nil && cs.State.Waiting.Reason != "" && cs.State.Waiting.Reason != "PodInitializing":
		return "Init:" + cs.State.Waiting.Reason
	}

	return "Init:" + strconv.Itoa(count) + "/" + strconv.Itoa(initCount)
}
