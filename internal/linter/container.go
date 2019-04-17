package linter

import (
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Docker image latest tag.
const imageTagLatest = "latest"

// Container represents a Container linter.
type Container struct {
	*Linter
}

// NewContainer returns a new container linter.
func NewContainer(l Loader, log *zerolog.Logger) *Container {
	return &Container{NewLinter(l, log)}
}

// Lint a Container.
func (c *Container) lint(co v1.Container, checkProbes bool) {
	c.checkImageTags(co)
	c.checkResources(co)
	if checkProbes {
		c.checkProbes(co)
	}
	c.checkNamedPorts(co)
}

func (c *Container) checkImageTags(co v1.Container) {
	tokens := strings.Split(co.Image, ":")
	if len(tokens) < 2 {
		c.addIssue(co.Name, ErrorLevel, "Untagged docker image in use")
		return
	}

	if tokens[1] == imageTagLatest {
		c.addIssue(co.Name, WarnLevel, "Image tagged `latest in use")
	}
}

func (c *Container) checkProbes(co v1.Container) {
	if co.LivenessProbe == nil && co.ReadinessProbe == nil {
		c.addIssue(co.Name, WarnLevel, "No probes defined")
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
		c.addIssue(co, InfoLevel, fmt.Sprintf("%s probe uses a port#, prefer a named port", kind))
	}
}

func (c *Container) checkResources(co v1.Container) {
	if len(co.Resources.Limits) == 0 && len(co.Resources.Requests) == 0 {
		c.addIssue(co.Name, WarnLevel, "No resources defined")
		return
	}

	if len(co.Resources.Requests) > 0 && len(co.Resources.Limits) == 0 {
		c.addIssue(co.Name, WarnLevel, "No resource limits defined")
	}
}

func (c *Container) checkNamedPorts(co v1.Container) {
	for _, p := range co.Ports {
		if len(p.Name) == 0 {
			c.addIssuef(co.Name, WarnLevel, "Unnamed port `%d", p.ContainerPort)
		}
	}
}

func (c *Container) checkUtilization(co v1.Container, cmx k8s.Metrics) {
	cpu, mem, _ := containerResources(co)
	c.checkMetrics(co.Name, cpu, mem, cmx.CurrentCPU, cmx.CurrentMEM)
}

func (c *Container) checkMetrics(co string, cpu, mem, ccpu, cmem resource.Quantity) {
	percCPU := ToPerc(toMC(ccpu), toMC(cpu))
	cpuLimit := int64(c.PodCPULimit())
	if percCPU >= cpuLimit {
		c.addIssuef(co, ErrorLevel, "CPU C:%s|R:%s reached user %d%% threshold (%d%%)", asMC(ccpu), asMC(cpu), cpuLimit, percCPU)
	}

	percMEM := ToPerc(toMB(cmem), toMB(mem))
	memLimit := int64(c.PodMEMLimit())
	if percMEM >= memLimit {
		c.addIssuef(co, ErrorLevel, "Memory C:%s|R:%s reached user %d%% threshold (%d%%)", asMB(cmem), asMB(mem), memLimit, percMEM)
	}
}

func containerResources(co v1.Container) (cpu, mem resource.Quantity, burstable bool) {
	req, limit := co.Resources.Requests, co.Resources.Limits

	switch {
	case len(req) != 0 && len(limit) != 0:
		cpu, mem = limit[v1.ResourceCPU], limit[v1.ResourceMemory]
		burstable = true
	case len(req) != 0:
		cpu, mem = req[v1.ResourceCPU], req[v1.ResourceMemory]
	case len(limit) != 0:
		cpu, mem = limit[v1.ResourceCPU], limit[v1.ResourceMemory]
	}

	return
}
