// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import _ "embed"

const (
	// DefaultFormat dumps report with color, emojis, the works.
	DefaultFormat = "standard"

	// JurassicFormat dumps report with dud fancy-ness.
	JurassicFormat = "jurassic"

	// YAMLFormat dumps report as YAML.
	YAMLFormat = "yaml"

	// JSONFormat dumps report as JSON.
	JSONFormat = "json"

	// HTMLFormat dumps report as HTML
	HTMLFormat = "html"

	// JunitFormat renders report as JUnit.
	JunitFormat = "junit"

	// ScoreFormat renders report as the value of the Score.
	ScoreFormat = "score"

	// PromFormat renders report to prom metrics.
	PromFormat = "prometheus"
)

//go:embed assets/report.html
var htmlReport string
