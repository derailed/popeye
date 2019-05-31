package sanitize

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// MegaByte represents a Mb.
const megaByte = 1024 * 1024

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
func asPerc(n int64) string {
	return fmt.Sprintf("%d%%", n)
}

// ToMCRatio computes millicore ratio.
func toMCRatio(q1, q2 resource.Quantity) int64 {
	if q2.IsZero() {
		return 0
	}
	v1, v2 := toMC(q1), toMC(q2)

	return int64(math.Round((float64(v1) / float64(v2)) * 100))
}

// ToMEMRatio computes mem MB ratio.
func toMEMRatio(q1, q2 resource.Quantity) int64 {
	if q2.IsZero() {
		return 0
	}
	v1, v2 := toMB(q1), toMB(q2)

	return int64(math.Round((float64(v1) / float64(v2)) * 100))
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
func containerResources(co v1.Container) (cpu, mem *resource.Quantity, burstable bool) {
	req, limit := co.Resources.Requests, co.Resources.Limits
	switch {
	case len(req) != 0 && len(limit) != 0:
		cpu, mem = req.Cpu(), req.Memory()
		burstable = true
	case len(req) != 0:
		cpu, mem = req.Cpu(), req.Memory()
	case len(limit) != 0:
		cpu, mem = limit.Cpu(), limit.Memory()
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
