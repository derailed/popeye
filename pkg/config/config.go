package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const defaultLintLevel = "ok"

// Config tracks Popeye configuration options.
type Config struct {
	Popeye    Popeye `yaml:"popeye"`
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
			return nil, err
		}
	}

	cfg.Flags = flags

	if flags.AllNamespaces != nil && *flags.AllNamespaces {
		var allNS string
		flags.Namespace = &allNS
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
	return c.Popeye.Allocations.CPU
}

// MEMResourceLimits returns memory over/under allocation thresholds.
func (c *Config) MEMResourceLimits() Allocations {
	return c.Popeye.Allocations.MEM
}

// NodeCPULimit returns the node cpu threshold if set otherwise the default.
func (c *Config) NodeCPULimit() float64 {
	l := c.Popeye.Node.Limits.CPU
	if l == 0 {
		return defaultCPULimit
	}
	return l
}

// PodCPULimit returns the pod cpu threshold if set otherwise the default.
func (c *Config) PodCPULimit() float64 {
	l := c.Popeye.Pod.Limits.CPU
	if l == 0 {
		return defaultCPULimit
	}
	return l
}

// ExcludedNode returns excluded nodes if any.
func (c *Config) ExcludedNode(n string) bool {
	return c.Popeye.Node.excluded(n)
}

// ExcludedService returns excluded services if any.
func (c *Config) ExcludedService(s string) bool {
	return c.Popeye.Service.excluded(s)
}

// ExcludedPod checks if a pod should be excluded from the scan.
func (c *Config) ExcludedPod(n string) bool {
	return c.Popeye.Pod.excluded(n)
}

// ExcludedNS checks if a namespace should be excluded from the scan.
func (c *Config) ExcludedNS(n string) bool {
	return c.Popeye.Namespace.excluded(n)
}

// RestartsLimit returns pod restarts limit.
func (c *Config) RestartsLimit() int {
	l := c.Popeye.Pod.Restarts
	if l == 0 {
		return defaultRestarts
	}
	return l
}

// PodMEMLimit returns the pod mem threshold if set otherwise the default.
func (c *Config) PodMEMLimit() float64 {
	l := c.Popeye.Pod.Limits.Memory
	if l == 0 {
		return defaultMEMLimit
	}
	return l
}

// NodeMEMLimit returns the pod mem threshold if set otherwise the default.
func (c *Config) NodeMEMLimit() float64 {
	l := c.Popeye.Node.Limits.Memory
	if l == 0 {
		return defaultMEMLimit
	}
	return l
}

// ----------------------------------------------------------------------------
// Helpers...

// IsSet checks if a string flag is set.
func isSet(s *string) bool {
	return s != nil && *s != ""
}
