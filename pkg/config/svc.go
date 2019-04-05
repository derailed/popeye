package config

// Service tracks service configurations.
type Service struct {
	Excludes `yaml:"exclude"`
}

// NewService create a new service configuration.
func newService() Service {
	return Service{}
}
