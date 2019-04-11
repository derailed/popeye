package config

var excludedNS = Excludes{"kube-public"}

// Namespace tracks namespace configurations.
type Namespace struct {
	Excludes `yaml:"exclude"`
}

// NewNamespace create a new namespace configuration.
func newNamespace() Namespace {
	return Namespace{Excludes: excludedNS}
}
