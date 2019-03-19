package linter

import (
	"math"
	"regexp"

	v1 "k8s.io/api/core/v1"
)

// Container represents a Container linter.
type Container struct {
	*Linter
}

// NewContainer returns a new linter.
func NewContainer() *Container {
	return &Container{new(Linter)}
}

// Lint a Container.
func (c *Container) Lint(co v1.Container) {
	c.checkImageTags(co)
	c.checkProbes(co)
	c.checkResources(co)
	c.checkNamedPorts(co)
}

var imageTagBruteRX = regexp.MustCompile(`\A(.+):(.+)\z`)

const imageTagLatest = "latest"

func (c *Container) checkImageTags(co v1.Container) {
	tokens := imageTagBruteRX.FindStringSubmatch(co.Image)
	if len(tokens) < 3 {
		c.addIssuef(WarnLevel, "No image tag was given on container `%s", co.Name)
		return
	}

	if tokens[2] == imageTagLatest {
		c.addIssuef(WarnLevel, "Lastest tag found on container `%s.", co.Name)
	}
}

func (c *Container) checkProbes(co v1.Container) {
	if co.LivenessProbe == nil {
		c.addIssuef(InfoLevel, "No liveness probe on container `%s", co.Name)
	}
	if co.ReadinessProbe == nil {
		c.addIssuef(InfoLevel, "No readiness probe on container `%s", co.Name)
	}
}

func (c *Container) checkResources(co v1.Container) {
	if len(co.Resources.Limits) == 0 && len(co.Resources.Requests) == 0 {
		c.addIssuef(InfoLevel, "No resources specified on container `%s", co.Name)
	}
}

func (c *Container) checkNamedPorts(co v1.Container) {
	for _, p := range co.Ports {
		if len(p.Name) == 0 {
			c.addIssuef(InfoLevel, "Unamed port found on container `%s", co.Name)
		}
	}
}

func (c *Container) checkUtilization(co v1.Container, cmx PodMetric) {
	cpu, mem := c.getLimits(co)
	c.checkMetrics(co.Name, cpu, mem, cmx.CurrentCPU(), cmx.CurrentMEM())
}

func (c *Container) checkMetrics(co string, cpu, mem, ccpu, cmem int64) {
	if cpu == 0 {
		return
	}
	percCPU := math.Round(float64(ccpu) / float64(cpu) * 100)
	if percCPU >= cpuPodLimit {
		c.addIssuef(WarnLevel, "CPU threshold reached on container `%s (%0.f%%)", co, percCPU)
	}

	if cmem == 0 {
		return
	}
	percMEM := math.Round(float64(cmem) / float64(mem) * 100)
	if percMEM >= memPodLimit {
		c.addIssuef(WarnLevel, "Memory threshold reached on container `%s (%0.f%%)", co, percMEM)
	}
}

func (c *Container) getLimits(co v1.Container) (cpu int64, mem int64) {
	req, limit := co.Resources.Requests, co.Resources.Limits
	if len(req) == 0 && len(limit) == 0 {
		return
	}

	if len(req) != 0 && len(limit) == 0 {
		lcpu := req[v1.ResourceCPU]
		cpu = lcpu.MilliValue()
		lmem := req[v1.ResourceMemory]
		if m, ok := lmem.AsInt64(); ok {
			mem = m
		}
		return
	}

	if len(limit) != 0 {
		lcpu := limit[v1.ResourceCPU]
		cpu = lcpu.MilliValue()
		lmem := limit[v1.ResourceMemory]
		if m, ok := lmem.AsInt64(); ok {
			mem = m
		}
	}
	return
}
