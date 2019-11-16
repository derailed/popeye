package sanitize

import (
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	qosBestEffort qos = iota
	qosBurstable
	qosGuaranteed

	// MegaByte represents a Mb.
	megaByte = 1024 * 1024
)

type qos = int

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

// In checks if a string is in a list of strings.
func in(ll []string, s string) bool {
	for _, l := range ll {
		if l == s {
			return true
		}
	}

	return false
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

// PodResources computes pod resouces as sum of containers allocations.
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
