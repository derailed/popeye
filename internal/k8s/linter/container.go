package linter

import (
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
