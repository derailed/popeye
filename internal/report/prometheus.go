package report

import (
	"fmt"
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

func prometheusMarshal(b *Builder, gtwy *config.PushGateway, cluster, namespace string) *push.Pusher {
	pusher := newPusher(gtwy)

	score.WithLabelValues(cluster, namespace, b.Report.Grade).Set(float64(b.Report.Score))
	errs.WithLabelValues(cluster, namespace).Set(float64(len(b.Report.Errors)))

	for _, section := range b.Report.Sections {
		for i, v := range section.Tally.counts {
			sanitizers.WithLabelValues(cluster, namespace, section.Title,
				strings.ToLower(indexToTally(i))).Set(float64(v))
		}
		sanitizersScore.WithLabelValues(cluster, namespace, section.Title).Set(float64(section.Tally.score))
	}
	return pusher
}

func newPusher(gtwy *config.PushGateway) *push.Pusher {
	registry := prometheus.NewRegistry()
	registry.MustRegister(score, errs, sanitizers, sanitizersScore)
	p := push.New(*gtwy.Address, "popeye").Gatherer(registry)
	if isSet(gtwy.BasicAuth.User) && isSet(gtwy.BasicAuth.Password) {
		fmt.Println("Using auth! ", *gtwy.BasicAuth.User, *gtwy.BasicAuth.Password)
		p = p.BasicAuth(*gtwy.BasicAuth.User, *gtwy.BasicAuth.Password)
	}

	return p
}

func isSet(s *string) bool {
	return s != nil && *s != ""
}
