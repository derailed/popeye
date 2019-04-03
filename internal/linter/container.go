package linter

import (
	"strings"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

// Docker image latest tag.
const imageTagLatest = "latest"

// Container represents a Container linter.
type Container struct {
	*Linter
}

// NewContainer returns a new container linter.
func NewContainer(c *k8s.Client, l *zerolog.Logger) *Container {
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
		c.addIssue(co.Name, ErrorLevel, "No probes found")
		return
	}

	if co.LivenessProbe == nil {
		c.addIssue(co.Name, WarnLevel, "No liveness probe")
	}
	if co.ReadinessProbe == nil {
		c.addIssue(co.Name, WarnLevel, "No readiness probe")
	}
}

func (c *Container) checkResources(co v1.Container) {
	if len(co.Resources.Limits) == 0 && len(co.Resources.Requests) == 0 {
		c.addIssue(co.Name, ErrorLevel, "No resources requests/limits found")
		return
	}

	if len(co.Resources.Requests) > 0 && len(co.Resources.Limits) == 0 {
		c.addIssue(co.Name, WarnLevel, "No resource limits found")
	}
}

func (c *Container) checkNamedPorts(co v1.Container) {
	for _, p := range co.Ports {
		if len(p.Name) == 0 {
			c.addIssuef(co.Name, WarnLevel, "Unamed port %d found", p.ContainerPort)
		}
	}
}

func (c *Container) checkUtilization(co v1.Container, cmx k8s.Metrics) {
	cpu, mem := c.getLimits(co)
	c.checkMetrics(co.Name, cpu, cmx.CurrentCPU, mem, int64(cmx.CurrentMEM))
}

func (c *Container) checkMetrics(co string, cpu, ccpu, mem, cmem int64) {
	percCPU := toPerc(float64(ccpu), float64(cpu))
	cpuLimit := c.client.Config.PodCPULimit()
	if percCPU >= cpuLimit {
		c.addIssuef(co, ErrorLevel, "CPU threshold reached %0.f%% (%0.f%%)", percCPU, cpuLimit)
	}

	percMEM := toPerc(float64(cmem), float64(mem))
	memLimit := c.client.Config.PodMEMLimit()
	if percMEM >= memLimit {
		c.addIssuef(co, ErrorLevel, "Memory threshold reached %0.f%% (%0.f%%)", percMEM, memLimit)
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
