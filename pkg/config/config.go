package config

import (
	"fmt"
	"io/ioutil"

	"github.com/derailed/popeye/internal/client"
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
	cfg := Config{Popeye: NewPopeye()}

	if isSet(flags.Spinach) {
		f, err := ioutil.ReadFile(*flags.Spinach)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(f, &cfg); err != nil {
			return nil, fmt.Errorf("Invalid spinach config file -- %w", err)
		}
	}
	cfg.Flags = flags

	if flags.Namespace != nil && *flags.Namespace == client.AllNamespaces {
		flags.Namespace = nil
	}
	if flags.AllNamespaces != nil && *flags.AllNamespaces {
		all := client.AllNamespaces
		flags.Namespace = &all
	}
	cfg.LintLevel = int(ToIssueLevel(flags.LintLevel))

	return &cfg, nil
}

// LinterLevel returns the current lint level.
func (c *Config) LinterLevel() int {
	return c.LintLevel
}

// Sections returns a collection of sanitizers categories.
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
	l := c.Node.Limits.CPU
	if l == 0 {
		return defaultCPULimit
	}
	return l
}

// PodCPULimit returns the pod cpu threshold if set otherwise the default.
func (c *Config) PodCPULimit() float64 {
	l := c.Pod.Limits.CPU
	if l == 0 {
		return defaultCPULimit
	}
	return l
}

// RestartsLimit returns pod restarts limit.
func (c *Config) RestartsLimit() int {
	l := c.Pod.Restarts
	if l == 0 {
		return defaultRestarts
	}
	return l
}

// PodMEMLimit returns the pod mem threshold if set otherwise the default.
func (c *Config) PodMEMLimit() float64 {
	l := c.Pod.Limits.Memory
	if l == 0 {
		return defaultMEMLimit
	}
	return l
}

// NodeMEMLimit returns the pod mem threshold if set otherwise the default.
func (c *Config) NodeMEMLimit() float64 {
	l := c.Node.Limits.Memory
	if l == 0 {
		return defaultMEMLimit
	}
	return l
}

func (c *Config) AllowedRegistries() []string {
	return c.Registries
}

// ----------------------------------------------------------------------------
// Helpers...

// IsSet checks if a string flag is set.
func isSet(s *string) bool {
	return s != nil && *s != ""
}
