// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"strconv"
	"strings"

	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rs/zerolog/log"
)

const namespace = "popeye"

// Metrics
var (
	sevGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "severity_total",
		Help:      "Popeye's severity scores totals.",
	},
		[]string{
			"cluster",
			"namespace",
			"severity",
		})

	codeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "code_total",
		Help:      "Popeye's report codes totals",
	},
		[]string{
			"cluster",
			"namespace",
			"linter",
			"code",
			"severity",
		})

	linterGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "linter_tally_total",
		Help:      "Popeye's linter tally totals",
	},
		[]string{
			"cluster",
			"linter",
			"severity",
		})

	errGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "report_errors_total",
		Help:      "Popeye's scan errors total.",
	},
		[]string{
			"cluster",
			"namespace",
		})

	scoreGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "cluster_score",
		Help:      "Popeye's scan cluster score.",
	},
		[]string{
			"cluster",
			"namespace",
			"grade",
		})

	reportGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "report_score",
		Help:      "Popeye's scan report score.",
	},
		[]string{
			"cluster",
			"namespace",
			"grade",
			"scan",
		})
)

func (b *Builder) promCollect(ns, scanReport string, codes rules.Glossary) {
	cc := b.Report.Sections.CodeTallies()
	cc.Compact()
	cc.Dump()

	cl := b.ClusterName
	scoreGauge.WithLabelValues(cl, ns, b.Report.Grade).Set(float64(b.Report.Score))
	reportGauge.WithLabelValues(cl, ns, b.Report.Grade, scanReport).Set(float64(b.Report.Score))
	errGauge.WithLabelValues(cl, ns).Set(float64(len(b.Report.Errors)))

	for linter, nss := range cc {
		for ns, st := range nss {
			for level, count := range st.Rollup(codes) {
				sevGauge.WithLabelValues(cl, ns, level.ToHumanLevel()).Add(float64(count))
			}
			for code, count := range st {
				cid, _ := strconv.Atoi(code)
				c := codes[rules.ID(cid)]
				codeGauge.WithLabelValues(cl, ns, linter, code, c.Severity.ToHumanLevel()).Add(float64(count))
			}
		}
	}
	for _, section := range b.Report.Sections {
		for i, v := range section.Tally.counts {
			linterGauge.WithLabelValues(cl, section.Title, strings.ToLower(indexToTally(i))).Add(float64(v))
		}
	}
}

func newPusher(gtwy *config.PushGateway, instance string) *push.Pusher {
	registry := prometheus.NewRegistry()
	registry.MustRegister(scoreGauge, errGauge, linterGauge, sevGauge, codeGauge, reportGauge)

	pusher := push.New(*gtwy.URL, "popeye").
		Gatherer(registry).
		Grouping("instance", instance)

	if config.IsStrSet(gtwy.BasicAuth.User) && config.IsStrSet(gtwy.BasicAuth.Password) {
		log.Debug().Msgf("Using basic auth: %s", *gtwy.BasicAuth.User)
		pusher = pusher.BasicAuth(*gtwy.BasicAuth.User, *gtwy.BasicAuth.Password)
	}

	return pusher
}
