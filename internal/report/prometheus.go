package report

import (
	"strings"

	"github.com/derailed/popeye/pkg/config"
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

func prometheusMarshal(b *Builder, address *string, level *string, lintdetail *string, cluster, namespace string) *push.Pusher {
	pusher := newPusher(address)

	score.WithLabelValues(cluster, namespace, b.Report.Grade).Set(float64(b.Report.Score))
	errs.WithLabelValues(cluster, namespace).Set(float64(len(b.Report.Errors)))

	for _, section := range b.Report.Sections {
		for i, v := range section.Tally.counts {
			detailedReport := ""
			keys := make([]string, 0, len(section.Outcome))
			if *level == VerboseOutputDetail {
				for k, cissues := range section.Outcome {
					for _, cissue := range cissues {
						if int(cissue.Level) == i && i >= int(config.ToIssueLevel(lintdetail)) && float64(v) > 0 {
							keys = append(keys, k+" "+cissue.Message)
						}
					}
					detailedReport = strings.Join(keys, ",")
				}
			}
			sanitizers.WithLabelValues(cluster, namespace, section.Title,
				strings.ToLower(indexToTally(i)), detailedReport).Set(float64(v))
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
