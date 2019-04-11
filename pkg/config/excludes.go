package config

// Excludes represents a lists items that should be excluded.
type Excludes []string

func (e Excludes) excluded(name string) bool {
	for _, n := range e {
		if n == name {
			return true
		}
	}
	return false
}
