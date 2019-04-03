package config

var excludedNS = Excludes{"kube-public"}

// Namespace tracks namespace configurations.
type Namespace struct {
	Excludes []string `yaml:"exclude"`
}

// NewNamespace create a new namespace configuration.
func newNamespace() Namespace {
	return Namespace{Excludes: excludedNS}
}

func (n Namespace) excluded(name string) bool {
	for _, ns := range n.Excludes {
		if ns == name {
			return true
		}
	}
	return false
}
