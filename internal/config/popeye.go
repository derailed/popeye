package config

import (
	"github.com/rs/zerolog"
)

// Popeye tracks Popeye configuration options.
type Popeye struct {
	Namespace Namespace `yaml:"namespace"`
	Node      Node      `yaml:"node"`
	Pod       Pod       `yaml:"pod"`

	LogLevel  zerolog.Level
	LintLevel int
}

// NewPopeye create a new Popeye configuration.
func NewPopeye() Popeye {
	return Popeye{
		LogLevel:  zerolog.DebugLevel,
		LintLevel: 1,
		Namespace: newNamespace(),
		Node:      newNode(),
		Pod:       newPod(),
	}
}
