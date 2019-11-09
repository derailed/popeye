package report

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const namespace = "popeye"

// Metrics
var (
	score = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "score",
		Help:      "Score of kubernetes cluster.",
	},
		[]string{
			"cluster",
			"namespace",
		})
	grade = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "grade",
		Help:      "Grade of kubernetes cluster. (1: A, 2: B, 3: C, 4: D, 5: E, 6: F)",
	},
		[]string{
			"cluster",
			"namespace",
			"grade",
		})
	sanitizersOk = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizers_ok",
		Help:      "Sanitizer ok level results for resource groups.",
	},
		[]string{
			"cluster",
			"namespace",
			"title",
		})
	sanitizersInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizers_info",
		Help:      "Sanitizer info level results for resource groups.",
	},
		[]string{
			"cluster",
			"namespace",
			"title",
		})
	sanitizersWarning = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizers_warning",
		Help:      "Sanitizer warning level results for resource groups.",
	},
		[]string{
			"cluster",
			"namespace",
			"title",
		})
	sanitizersError = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizers_error",
		Help:      "Sanitizer error level results for resource groups.",
	},
		[]string{
			"cluster",
			"namespace",
			"title",
		})
	sanitizersScore = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizers_score",
		Help:      "Sanitizer score results for resource groups.",
	},
		[]string{
			"cluster",
			"namespace",
			"title",
		})
	errs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "errors",
		Help:      "Errors while sanitizing the cluster.",
	},
		[]string{
			"cluster",
			"namespace",
		})
)

func prometheusMarshal(b *Builder, address *string, cluster, namespace string) *push.Pusher {
	pusher := newPusher(address)

	score.WithLabelValues(cluster, namespace).Set(float64(b.Report.Score))
	grade.WithLabelValues(cluster, namespace, b.Report.Grade).Set(float64(gradeToNumber(b.Report.Grade)))
	errs.WithLabelValues(cluster, namespace).Set(float64(len(b.Report.Errors)))

	for _, section := range b.Report.Sections {
		for i, v := range section.Tally.counts {
			switch i {
			case 0:
				sanitizersOk.WithLabelValues(cluster, namespace, section.Title).Set(float64(v))
			case 1:
				sanitizersInfo.WithLabelValues(cluster, namespace, section.Title).Set(float64(v))
			case 2:
				sanitizersWarning.WithLabelValues(cluster, namespace, section.Title).Set(float64(v))
			case 3:
				sanitizersError.WithLabelValues(cluster, namespace, section.Title).Set(float64(v))
			}
		}
		sanitizersScore.WithLabelValues(cluster, namespace, section.Title).Set(float64(section.Tally.score))
	}

	return pusher
}

func newPusher(address *string) *push.Pusher {
	registry := prometheus.NewRegistry()
	registry.MustRegister(score, grade, errs,
		sanitizersOk, sanitizersWarning, sanitizersInfo, sanitizersError, sanitizersScore)
	return push.New(*address, "popeye").Gatherer(registry)
}

func gradeToNumber(grade string) int {
	switch grade {
	case "A":
		return 1
	case "B":
		return 2
	case "C":
		return 3
	case "D":
		return 4
	case "E":
		return 5
	case "F":
		return 6
	default:
		return 6
	}
}
