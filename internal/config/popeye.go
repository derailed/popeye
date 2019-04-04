package config

import (
	"github.com/rs/zerolog"
)

// Popeye tracks Popeye configuration options.
type Popeye struct {
	Namespace Namespace `yaml:"namespace"`
	Node      Node      `yaml:"node"`
	Pod       Pod       `yaml:"pod"`
	Service   Service   `yaml:"service"`

	LogLevel  zerolog.Level
	LintLevel Level
}

// NewPopeye create a new Popeye configuration.
func NewPopeye() Popeye {
	return Popeye{
		LogLevel:  zerolog.DebugLevel,
		Namespace: newNamespace(),
		Node:      newNode(),
		Pod:       newPod(),
		Service:   newService(),
	}
}
