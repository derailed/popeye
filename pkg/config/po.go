// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

const defaultRestarts = 5

// Pod tracks pod configurations.
type Pod struct {
	Restarts int    `yaml:"restarts"`
	Limits   Limits `yaml:"limits"`
}

// NewPod create a new pod configuration.
func newPod() Pod {
	return Pod{
		Restarts: defaultRestarts,
		Limits: Limits{
			CPU:    defaultCPULimit,
			Memory: defaultMEMLimit,
		},
	}
}
