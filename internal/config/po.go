package config

// Pod tracks pod configurations.
type Pod struct {
	Limits  Limits   `yaml:"limits`
	Exclude Excludes `yaml:"exclude"`
}

// NewPod create a new pod configuration.
func newPod() Pod {
	return Pod{
		Limits: Limits{
			CPU:    defaultCPULimit,
			Memory: defaultMEMLimit,
		},
	}
}
