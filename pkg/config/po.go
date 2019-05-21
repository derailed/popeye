package config

const defaultRestarts = 5

// Pod tracks pod configurations.
type Pod struct {
	Restarts int    `yaml:"restarts"`
	Limits   Limits `yaml:"limits"`
	Excludes `yaml:"exclude"`
}

// NewPod create a new pod configuration.
func newPod() Pod {
	return Pod{
		Restarts: defaultRestarts,
		Limits: Limits{
			CPU:    defaultCPULimit,
			Memory: defaultMEMLimit,
		},
	}
}
