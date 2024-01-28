// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

const (
	defaultCPULimit = 80 // percentage
	defaultMEMLimit = 80 // percentage
)

// Limits tracks cpu and mem limits.
type Limits struct {
	CPU    float64 `yaml:"cpu"`
	Memory float64 `yam:"memory"`
}

// Node tracks node configurations.
type Node struct {
	Limits Limits `yaml:"limits"`
}

// NewNode create a new node configuration.
func newNode() Node {
	return Node{
		Limits: Limits{
			CPU:    defaultCPULimit,
			Memory: defaultMEMLimit,
		},
	}
}
