package report

import (
	"strings"

	"github.com/derailed/popeye/internal/issues"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const namespace = "popeye"

// Metrics
var (
	score = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "cluster_score_total",
		Help:      "Popeye's sanitizers overall cluster score.",
	},
		[]string{
			"cluster",
			"namespace",
			"grade",
		})
	sanitizers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizer_reports_count",
		Help:      "Popeye's sanitizer reports for resource group.",
	},
		[]string{
			"cluster",
			"namespace",
			"resource",
			"level",
			"issues",
		})
	sanitizersScore = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sanitizer_score_total",
		Help:      "Popeye's sanitizer score for resource group.",
	},
		[]string{
			"cluster",
			"namespace",
			"resource",
		})
	errs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "errors_total",
		Help:      "Popeye's sanitizers errors.",
	},
		[]string{
			"cluster",
			"namespace",
		})
)

func prometheusMarshal(b *Builder, address *string, cluster, namespace string) *push.Pusher {
	pusher := newPusher(address)

	score.WithLabelValues(cluster, namespace, b.Report.Grade).Set(float64(b.Report.Score))
	errs.WithLabelValues(cluster, namespace).Set(float64(len(b.Report.Errors)))

	for _, section := range b.Report.Sections {
		for i, v := range section.Tally.counts {
			issuesReport := ""
			keys := make([]string, 0, len(section.Outcome))
			for k, ci := range section.Outcome {
				for _, gg := range ci {
					if int(gg.Level) == i && i > int(issues.InfoLevel) && float64(v) > 0 {
						keys = append(keys, k+" "+gg.Message)
					}
				}
				issuesReport = strings.Join(keys, ",")
			}
			sanitizers.WithLabelValues(cluster, namespace, section.Title,
				strings.ToLower(indexToTally(i)), issuesReport).Set(float64(v))
		}
		sanitizersScore.WithLabelValues(cluster, namespace, section.Title).Set(float64(section.Tally.score))
	}
	return pusher
}

func newPusher(address *string) *push.Pusher {
	registry := prometheus.NewRegistry()
	registry.MustRegister(score, errs, sanitizers, sanitizersScore)
	return push.New(*address, "popeye").Gatherer(registry)
}
