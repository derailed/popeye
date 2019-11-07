package report

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const namespace = "popeye"

// Metrics
var (
	score = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "score",
		Help:      "Score of kubernetes cluster.",
	})
	grade = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "grade",
		Help:      "Grade of kubernetes cluster. (1: A, 2: B, 3: C, 4: D, 5: E, 6: F)",
	})
	sanitizersOk = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizers_ok",
		Help:      "Sanitizer ok level results for resource groups.",
	},
		[]string{
			"title",
		})
	sanitizersInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizers_info",
		Help:      "Sanitizer info level results for resource groups.",
	},
		[]string{
			"title",
		})
	sanitizersWarning = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizers_warning",
		Help:      "Sanitizer warning level results for resource groups.",
	},
		[]string{
			"title",
		})
	sanitizersError = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizers_error",
		Help:      "Sanitizer error level results for resource groups.",
	},
		[]string{
			"title",
		})
	sanitizersScore = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizers_score",
		Help:      "Sanitizer score results for resource groups.",
	},
		[]string{
			"title",
		})
	errs = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "errors",
		Help:      "Errors while sanitizing the cluster.",
	})
)

func prometheusMarshal(b *Builder, address *string) *push.Pusher {
	pusher := newPusher(address)

	score.Set(float64(b.Report.Score))
	grade.Set(float64(gradeToNumber(b.Report.Grade)))
	errs.Set(float64(len(b.Report.Errors)))

	for _, section := range b.Report.Sections {
		for i, v := range section.Tally.counts {
			switch i {
			case 0:
				sanitizersOk.WithLabelValues(section.Title).Set(float64(v))
			case 1:
				sanitizersInfo.WithLabelValues(section.Title).Set(float64(v))
			case 2:
				sanitizersWarning.WithLabelValues(section.Title).Set(float64(v))
			case 3:
				sanitizersError.WithLabelValues(section.Title).Set(float64(v))
			}
		}
		sanitizersScore.WithLabelValues(section.Title).Set(float64(section.Tally.score))
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
