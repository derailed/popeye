package report

import (
	"os"
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
			"target",
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

func prometheusMarshal(b *Builder, t *os.File, f *config.Flags, cluster, namespace string) *push.Pusher {

	pusher := newPusher(f.PushGatewayAddress)
	outputTargetName := ""

	if *f.Save {
		outputTargetName = t.Name()
	}

	score.WithLabelValues(cluster, namespace, b.Report.Grade).Set(float64(b.Report.Score))
	errs.WithLabelValues(cluster, namespace).Set(float64(len(b.Report.Errors)))

	for _, section := range b.Report.Sections {
		for i, v := range section.Tally.counts {
			detailedReport := getIssues(section, f, i, v)
			sanitizers.WithLabelValues(cluster, namespace, section.Title,
				strings.ToLower(indexToTally(i)), detailedReport, outputTargetName).Set(float64(v))
		}
		sanitizersScore.WithLabelValues(cluster, namespace, section.Title).Set(float64(section.Tally.score))
	}
	return pusher
}

func getIssues(section Section, f *config.Flags, index int, value int) string {
	if *f.OutputDetail != VerboseOutputDetail {
		return ""
	}
	lintdetail := f.LintLevel
	detailedReport := ""
	keys := make([]string, 0, len(section.Outcome))
	for k, cissues := range section.Outcome {
		for _, cissue := range cissues {
			if int(cissue.Level) == index && index >= int(config.ToIssueLevel(lintdetail)) && float64(value) > 0 {
				keys = append(keys, k+" "+cissue.Message)
			}
		}
		detailedReport = strings.Join(keys, ",")
	}
	return detailedReport
}

func newPusher(address *string) *push.Pusher {
	registry := prometheus.NewRegistry()
	registry.MustRegister(score, errs, sanitizers, sanitizersScore)
	return push.New(*address, "popeye").Gatherer(registry)
}
