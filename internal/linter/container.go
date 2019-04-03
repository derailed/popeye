package linter

import (
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Docker image latest tag.
const imageTagLatest = "latest"

// Container represents a Container linter.
type Container struct {
	*Linter
}

// NewContainer returns a new container linter.
func NewContainer(c Client, l *zerolog.Logger) *Container {
	return &Container{newLinter(c, l)}
}

// Lint a Container.
func (c *Container) lint(co v1.Container) {
	c.checkImageTags(co)
	c.checkProbes(co)
	c.checkResources(co)
	c.checkNamedPorts(co)
}

func (c *Container) checkImageTags(co v1.Container) {
	tokens := strings.Split(co.Image, ":")
	if len(tokens) < 2 {
		c.addIssue(co.Name, ErrorLevel, "Non tagged image in use")
		return
	}

	if tokens[1] == imageTagLatest {
		c.addIssue(co.Name, WarnLevel, "Image tagged :latest in use")
	}
}

func (c *Container) checkProbes(co v1.Container) {
	if co.LivenessProbe == nil && co.ReadinessProbe == nil {
		c.addIssue(co.Name, ErrorLevel, "No probes defined")
		return
	}

	if co.LivenessProbe == nil {
		c.addIssue(co.Name, WarnLevel, "No liveness probe")
	}
	c.checkNamedProbe(co.Name, co.LivenessProbe, true)

	if co.ReadinessProbe == nil {
		c.addIssue(co.Name, WarnLevel, "No readiness probe")
	}
	c.checkNamedProbe(co.Name, co.ReadinessProbe, false)
}

func (c *Container) checkNamedProbe(co string, p *v1.Probe, liveness bool) {
	if p == nil || p.Handler.HTTPGet == nil {
		return
	}

	kind := "Readiness"
	if liveness {
		kind = "Liveness"
	}
	if p.Handler.HTTPGet != nil && p.Handler.HTTPGet.Port.Type == intstr.Int {
		c.addIssue(co, InfoLevel, fmt.Sprintf("%sProbe uses a port#, prefer a named port", kind))
	}
}

func (c *Container) checkResources(co v1.Container) {
	if len(co.Resources.Limits) == 0 && len(co.Resources.Requests) == 0 {
		c.addIssue(co.Name, ErrorLevel, "No resources defined")
		return
	}

	if len(co.Resources.Requests) > 0 && len(co.Resources.Limits) == 0 {
		c.addIssue(co.Name, WarnLevel, "No resource limits defined")
	}
}

func (c *Container) checkNamedPorts(co v1.Container) {
	for _, p := range co.Ports {
		if len(p.Name) == 0 {
			c.addIssuef(co.Name, WarnLevel, "Unamed port `%d", p.ContainerPort)
		}
	}
}

func (c *Container) checkUtilization(co v1.Container, cmx k8s.Metrics) {
	cpu, mem := c.getLimits(co)
	c.checkMetrics(co.Name, cpu, cmx.CurrentCPU, mem, int64(cmx.CurrentMEM))
}

func (c *Container) checkMetrics(co string, cpu, ccpu, mem, cmem int64) {
	percCPU := toPerc(float64(ccpu), float64(cpu))
	cpuLimit := c.client.PodCPULimit()
	if percCPU >= cpuLimit {
		c.addIssuef(co, ErrorLevel, "CPU threshold (%0.f%%) reached `%0.f%%", cpuLimit, percCPU)
	}

	percMEM := toPerc(float64(cmem), float64(mem))
	memLimit := c.client.PodMEMLimit()
	if percMEM >= memLimit {
		c.addIssuef(co, ErrorLevel, "Memory threshold (%0.f%%) reached `%0.f%%", memLimit, percMEM)
	}
}

func (c *Container) getLimits(co v1.Container) (cpu, mem int64) {
	req, limit := co.Resources.Requests, co.Resources.Limits

	switch {
	case len(req) == 0 && len(limit) == 0:
	case len(req) != 0 && len(limit) == 0:
		lcpu := req[v1.ResourceCPU]
		cpu = lcpu.MilliValue()
		lmem := req[v1.ResourceMemory]
		if m, ok := lmem.AsInt64(); ok {
			mem = m
		}
	case len(limit) != 0:
		lcpu := limit[v1.ResourceCPU]
		cpu = lcpu.MilliValue()
		lmem := limit[v1.ResourceMemory]
		if m, ok := lmem.AsInt64(); ok {
			mem = m
		}
	}
	return
}
