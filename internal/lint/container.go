// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"strings"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Docker image latest tag.
const imageTagLatest = "latest"

const defaultRegistry = "docker.io"

type (
	// LimitCollector represents a collector with resource limits.
	LimitCollector interface {
		Collector
		PodLimiter
		ContainerRestrictor
	}

	// Container represents a Container linter.
	Container struct {
		LimitCollector
		fqn string
	}
)

// NewContainer returns a new instance.
func NewContainer(fqn string, c LimitCollector) *Container {
	return &Container{fqn: fqn, LimitCollector: c}
}

func (c *Container) sanitize(ctx context.Context, co v1.Container, checkProbes bool) {
	ctx = internal.WithGroup(ctx, types.NewGVR("containers"), co.Name)
	c.checkImageTags(ctx, co.Image)
	if c.allowedRegistryListExists() {
		c.checkImageRegistry(ctx, co.Image)
	}
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

func (c *Container) checkImageRegistry(ctx context.Context, image string) {
	registries := c.LimitCollector.AllowedRegistries()
	tokens := strings.Split(image, "/")

	if len(tokens) == 1 {
		tokens[0] = defaultRegistry
	}

	for i := 0; i < len(registries); i++ {
		if tokens[0] == registries[i] {
			return
		}
	}

	c.AddSubCode(ctx, 113, image)
}

func (c *Container) checkProbes(ctx context.Context, co v1.Container) {
	if co.LivenessProbe == nil && co.ReadinessProbe == nil {
		c.AddSubCode(ctx, 102)
		return
	}
	if co.LivenessProbe == nil {
		c.AddSubCode(ctx, 103)
	} else {
		c.checkNamedProbe(ctx, co.LivenessProbe, true)
	}
	if co.ReadinessProbe == nil {
		c.AddSubCode(ctx, 104)
	} else {
		c.checkNamedProbe(ctx, co.ReadinessProbe, false)
	}
}

func (c *Container) checkNamedProbe(ctx context.Context, p *v1.Probe, liveness bool) {
	if p == nil || p.ProbeHandler.HTTPGet == nil {
		return
	}
	kind := "Readiness"
	if liveness {
		kind = "Liveness"
	}
	if p.ProbeHandler.HTTPGet != nil && p.ProbeHandler.HTTPGet.Port.Type == intstr.Int {
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

func (c *Container) checkUtilization(ctx context.Context, co v1.Container, cmx client.Metrics) {
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

func (c *Container) allowedRegistryListExists() bool {
	return len(c.LimitCollector.AllowedRegistries()) > 0
}
