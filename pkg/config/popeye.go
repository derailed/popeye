package config

// Popeye tracks Popeye configuration options.
type Popeye struct {
	Namespace Namespace `yaml:"namespace"`
	Node      Node      `yaml:"node"`
	Pod       Pod       `yaml:"pod"`
	Service   Service   `yaml:"service"`
}

// NewPopeye create a new Popeye configuration.
func NewPopeye() Popeye {
	return Popeye{
		Namespace: newNamespace(),
		Node:      newNode(),
		Pod:       newPod(),
		Service:   newService(),
	}
}
