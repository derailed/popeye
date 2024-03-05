// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import (
	"fmt"
	"os"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/pkg/config/json"
	"github.com/derailed/popeye/types"
	"gopkg.in/yaml.v2"
)

const defaultLintLevel = "ok"

// Config tracks Popeye configuration options.
type Config struct {
	Popeye    `yaml:"popeye"`
	Flags     *Flags
	LintLevel int
}

// NewConfig create a new Popeye configuration.
func NewConfig(flags *Flags) (*Config, error) {
	cfg := Config{
		Popeye: NewPopeye(),
	}

	if isSet(flags.Spinach) {
		bb, err := os.ReadFile(*flags.Spinach)
		if err != nil {
			return nil, err
		}
		if err := json.NewValidator().Validate(json.SpinachSchema, bb); err != nil {
			return nil, fmt.Errorf("validation failed for %q: %w", *flags.Spinach, err)
		}
		if err := yaml.Unmarshal(bb, &cfg); err != nil {
			return nil, fmt.Errorf("Invalid spinach config file -- %w", err)
		}
	}
	cfg.Flags = flags

	if flags.Namespace != nil && *flags.Namespace == client.AllNamespaces {
		flags.Namespace = nil
	}
	if flags.AllNamespaces != nil && *flags.AllNamespaces {
		all := client.NamespaceAll
		flags.Namespace = &all
	}
	cfg.LintLevel = int(rules.ToIssueLevel(flags.LintLevel))

	return &cfg, nil
}

func (c *Config) Match(s rules.Spec) bool {
	return c.Popeye.Match(s)
}

func (c *Config) ExcludeFQN(gvr types.GVR, fqn string, cos []string) bool {
	return c.Popeye.Match(rules.Spec{
		GVR:        gvr,
		FQN:        fqn,
		Containers: cos,
	})
}

func (c *Config) ExcludeContainer(gvr types.GVR, fqn, co string) bool {
	return c.Popeye.Match(rules.Spec{
		GVR:        gvr,
		FQN:        fqn,
		Containers: []string{co},
	})
}

// Sections tracks a collection of internal.
func (c *Config) Sections() []string {
	if c.Flags.Sections != nil {
		return *c.Flags.Sections
	}

	return []string{}
}

// CPUResourceLimits returns memory over/under allocation thresholds.
func (c *Config) CPUResourceLimits() Allocations {
	return c.CPU
}

// MEMResourceLimits returns memory over/under allocation thresholds.
func (c *Config) MEMResourceLimits() Allocations {
	return c.MEM
}

// NodeCPULimit returns the node cpu threshold if set otherwise the default.
func (c *Config) NodeCPULimit() float64 {
	l := c.Resources.Node.Limits.CPU
	if l == 0 {
		return defaultCPULimit
	}
	return l
}

// PodCPULimit returns the pod cpu threshold if set otherwise the default.
func (c *Config) PodCPULimit() float64 {
	l := c.Resources.Pod.Limits.CPU
	if l == 0 {
		return defaultCPULimit
	}
	return l
}

// RestartsLimit returns pod restarts limit.
func (c *Config) RestartsLimit() int {
	l := c.Resources.Pod.Restarts
	if l == 0 {
		return defaultRestarts
	}
	return l
}

// PodMEMLimit returns the pod mem threshold if set otherwise the default.
func (c *Config) PodMEMLimit() float64 {
	l := c.Resources.Pod.Limits.Memory
	if l == 0 {
		return defaultMEMLimit
	}
	return l
}

// NodeMEMLimit returns the pod mem threshold if set otherwise the default.
func (c *Config) NodeMEMLimit() float64 {
	l := c.Resources.Node.Limits.Memory
	if l == 0 {
		return defaultMEMLimit
	}
	return l
}

// AllowedRegistries tracks allowed docker registries.
func (c *Config) AllowedRegistries() []string {
	return c.Registries
}

// ----------------------------------------------------------------------------
// Helpers...

func isSet(s *string) bool {
	return s != nil && *s != ""
}
