package config

var excludedNS = Excludes{"kube-public"}

// ----------------------------------------------------------------------------

// Excludes lists items that should be excluded.
type Excludes []string

func (e Excludes) excluded(name string) bool {
	for _, n := range e {
		if n == name {
			return true
		}
	}
	return false
}

// ----------------------------------------------------------------------------

// Namespace tracks namespace configurations.
type Namespace struct {
	Excludes `yaml:"exclude"`
}

// NewNamespace create a new namespace configuration.
func newNamespace() Namespace {
	return Namespace{Excludes: excludedNS}
}
