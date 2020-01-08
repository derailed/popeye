package sanitize

import (
	"context"
	"strings"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/k8s"
	v1 "k8s.io/api/core/v1"
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

// NewContainer returns a new sanitizer.
func NewContainer(fqn string, c LimitCollector) *Container {
	return &Container{fqn: fqn, LimitCollector: c}
}

func (c *Container) sanitize(ctx context.Context, co v1.Container, checkProbes bool) {
	ctx = internal.WithFQN(ctx, c.fqn)
	ctx = internal.WithGroup(ctx, co.Name)
	c.checkImageTags(ctx, co.Image)
	c.checkResources(ctx, co)
	if checkProbes {
		c.checkProbes(ctx, co)
	}
	c.checkNamedPorts(ctx, co)
}

func (c *Container) checkImageTags(ctx context.Context, image string) {
	tokens := strings.Split(image, ":")
	if len(tokens) < 2 {
		c.AddSubCode(ctx, 100)
		return
	}

	if tokens[1] == imageTagLatest {
		c.AddSubCode(ctx, 101)
	}
}

func (c *Container) checkProbes(ctx context.Context, co v1.Container) {
	if co.LivenessProbe == nil && co.ReadinessProbe == nil {
		c.AddSubCode(ctx, 102)
		return
	}

	if co.LivenessProbe == nil {
		c.AddSubCode(ctx, 103)
	}
	c.checkNamedProbe(ctx, co.LivenessProbe, true)

	if co.ReadinessProbe == nil {
		c.AddSubCode(ctx, 104)
	}
	c.checkNamedProbe(ctx, co.ReadinessProbe, false)
}

func (c *Container) checkNamedProbe(ctx context.Context, p *v1.Probe, liveness bool) {
	if p == nil || p.Handler.HTTPGet == nil {
		return
	}
	kind := "Readiness"
	if liveness {
		kind = "Liveness"
	}
	if p.Handler.HTTPGet != nil && p.Handler.HTTPGet.Port.Type == intstr.Int {
		c.AddSubCode(ctx, 105, kind)
	}
}

func (c *Container) checkResources(ctx context.Context, co v1.Container) {
	if len(co.Resources.Limits) == 0 && len(co.Resources.Requests) == 0 {
		c.AddSubCode(ctx, 106)
		return
	}

	if len(co.Resources.Requests) > 0 && len(co.Resources.Limits) == 0 {
		c.AddSubCode(ctx, 107)
	}
}

func (c *Container) checkNamedPorts(ctx context.Context, co v1.Container) {
	for _, p := range co.Ports {
		if len(p.Name) == 0 {
			c.AddSubCode(ctx, 108, p.ContainerPort)
		}
	}
}

func (c *Container) checkUtilization(ctx context.Context, co v1.Container, cmx k8s.Metrics) {
	cpu, mem, qos := containerResources(co)
	if cpu != nil && mem != nil {
		ccpu, cmem := cmx.CurrentCPU, cmx.CurrentMEM
		list := v1.ResourceList{v1.ResourceCPU: *cpu, v1.ResourceMemory: *mem}
		cList := v1.ResourceList{v1.ResourceCPU: ccpu, v1.ResourceMemory: cmem}
		c.checkMetrics(ctx, qos, list, cList)
	}
}

func (c *Container) checkMetrics(ctx context.Context, qos qos, list, clist v1.ResourceList) {
	cpu, mem := list.Cpu(), list.Memory()
	ccpu, cmem := clist.Cpu(), clist.Memory()
	percCPU, cpuLimit := ToPerc(toMC(*ccpu), toMC(*cpu)), int64(c.PodCPULimit())
	percMEM, memLimit := ToPerc(toMB(*cmem), toMB(*mem)), int64(c.PodMEMLimit())

	switch qos {
	case qosBurstable:
		if percCPU > cpuLimit {
			c.AddSubCode(ctx, 109, asMC(*ccpu), asMC(*cpu), cpuLimit, percCPU)
		}
		if percMEM > memLimit {
			c.AddSubCode(ctx, 110, asMB(*cmem), asMB(*mem), memLimit, percMEM)
		}
	case qosGuaranteed:
		if percCPU > cpuLimit {
			c.AddSubCode(ctx, 111, asMC(*ccpu), asMC(*cpu), cpuLimit, percCPU)
		}
		if percMEM > memLimit {
			c.AddSubCode(ctx, 112, asMB(*cmem), asMB(*mem), memLimit, percMEM)
		}
	}
}
