package sanitize

import (
	"strings"

	"github.com/derailed/popeye/internal/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Docker image latest tag.
const imageTagLatest = "latest"

type (
	// LimitCollector represents a collector with resource limits.
	LimitCollector interface {
		Collector
		PodLimiter
	}

	// Container represents a Container linter.
	Container struct {
		LimitCollector
		fqn string
	}
)

// NewContainer returns a new container linter.
func NewContainer(fqn string, c LimitCollector) *Container {
	return &Container{fqn: fqn, LimitCollector: c}
}

// Lint a Container.
func (c *Container) sanitize(co v1.Container, checkProbes bool) {
	c.checkImageTags(co.Name, co.Image)
	c.checkResources(co)
	if checkProbes {
		c.checkProbes(co)
	}
	c.checkNamedPorts(co)
}

func (c *Container) checkImageTags(name, image string) {
	tokens := strings.Split(image, ":")
	if len(tokens) < 2 {
		c.AddSubError(c.fqn, name, "Untagged docker image in use")
		return
	}

	if tokens[1] == imageTagLatest {
		c.AddSubWarn(c.fqn, name, "Image tagged `latest in use")
	}
}

func (c *Container) checkProbes(co v1.Container) {
	if co.LivenessProbe == nil && co.ReadinessProbe == nil {
		c.AddSubWarn(c.fqn, co.Name, "No probes defined")
		return
	}

	if co.LivenessProbe == nil {
		c.AddSubWarn(c.fqn, co.Name, "No liveness probe")
	}
	c.checkNamedProbe(co.Name, co.LivenessProbe, true)

	if co.ReadinessProbe == nil {
		c.AddSubWarn(c.fqn, co.Name, "No readiness probe")
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
		c.AddSubInfof(c.fqn, co, "%s probe uses a port#, prefer a named port", kind)
	}
}

func (c *Container) checkResources(co v1.Container) {
	if len(co.Resources.Limits) == 0 && len(co.Resources.Requests) == 0 {
		c.AddSubWarn(c.fqn, co.Name, "No resources defined")
		return
	}

	if len(co.Resources.Requests) > 0 && len(co.Resources.Limits) == 0 {
		c.AddSubWarn(c.fqn, co.Name, "No resource limits defined")
	}
}

func (c *Container) checkNamedPorts(co v1.Container) {
	for _, p := range co.Ports {
		if len(p.Name) == 0 {
			c.AddSubWarnf(c.fqn, co.Name, "Unnamed port `%d", p.ContainerPort)
		}
	}
}

func (c *Container) checkUtilization(co v1.Container, cmx k8s.Metrics) {
	cpu, mem, _ := containerResources(co)
	if cpu != nil && mem != nil {
		c.checkMetrics(co.Name, *cpu, *mem, cmx.CurrentCPU, cmx.CurrentMEM)
	}
}

func (c *Container) checkMetrics(co string, cpu, mem, ccpu, cmem resource.Quantity) {
	percCPU := ToPerc(toMC(ccpu), toMC(cpu))
	cpuLimit := int64(c.PodCPULimit())
	if percCPU >= cpuLimit {
		c.AddSubErrorf(c.fqn, co, "CPU C:%s|R:%s reached user %d%% threshold (%d%%)", asMC(ccpu), asMC(cpu), cpuLimit, percCPU)
	}

	percMEM := ToPerc(toMB(cmem), toMB(mem))
	memLimit := int64(c.PodMEMLimit())
	if percMEM >= memLimit {
		c.AddSubErrorf(c.fqn, co, "Memory C:%s|R:%s reached user %d%% threshold (%d%%)", asMB(cmem), asMB(mem), memLimit, percMEM)
	}
}
