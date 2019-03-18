package linter

import (
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

// Container checks
// + image labels
// + port names

// Lint a Container.
func (c *Container) Lint(co v1.Container) {
	c.checkProbes(co)
	c.checkResources(co)
}

func (c *Container) checkProbes(co v1.Container) {
	if co.LivenessProbe == nil {
		c.addIssuef(InfoLevel, "%s container has no liveness probe", co.Name)
	}
	if co.ReadinessProbe == nil {
		c.addIssuef(InfoLevel, "%s container has no readiness probe", co.Name)
	}
}

func (c *Container) checkResources(co v1.Container) {
	if len(co.Resources.Limits) == 0 && len(co.Resources.Requests) == 0 {
		c.addIssuef(InfoLevel, "%s had no resource limits|requests specified.", co.Name)
	}
}
