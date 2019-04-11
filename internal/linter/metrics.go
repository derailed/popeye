package linter

const megaByte = 1024 * 1024

type (
	// Metrics represent an aggregation of all pod containers metrics.
	Metrics struct {
		CurrentCPU int64
		CurrentMEM float64
	}

	// NodeMetrics describes raw node metrics.
	NodeMetrics struct {
		CurrentCPU int64
		CurrentMEM float64
		AvailCPU   int64
		AvailMEM   float64
		TotalCPU   int64
		TotalMEM   float64
	}

	// NodesMetrics tracks usage metrics per nodes.
	NodesMetrics map[string]NodeMetrics

	// PodsMetrics tracks usage metrics per pods.
	PodsMetrics map[string]ContainerMetrics

	// ContainerMetrics tracks container metrics
	ContainerMetrics map[string]Metrics
)

// Empty checks if we have any metrics.
func (n NodeMetrics) Empty() bool {
	return n == NodeMetrics{}
}

// Empty checks if we have any metrics.
func (m Metrics) Empty() bool {
	return m == Metrics{}
}

// ----------------------------------------------------------------------------
// Helpers...

func asMi(v int64) float64 {
	return float64(v) / megaByte
}
