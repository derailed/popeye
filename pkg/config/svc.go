package config

var excludedSVC = Excludes{"default/kubernetes"}

// Service tracks service configurations.
type Service struct {
	Excludes `yaml:"exclude"`
}

// NewService create a new service configuration.
func newService() Service {
	return Service{Excludes: excludedSVC}
}
