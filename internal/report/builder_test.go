package report

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestBuilderJunit(t *testing.T) {
	b, ta := NewBuilder(internal.NewAliases()), NewTally()
	o := issues.Outcome{
		"blee": issues.Issues{
			issues.New(issues.Root, config.OkLevel, "Blah"),
		},
	}

	ta.Rollup(o)
	b.AddSection("fred", o, ta)
	b.AddError(errors.New("boom"))
	s, err := b.ToJunit(config.OkLevel)

	assert.Nil(t, err)
	assert.Equal(t, reportJunit, s)
}

func TestBuilderYAML(t *testing.T) {
	b, ta := NewBuilder(internal.NewAliases()), NewTally()
	o := issues.Outcome{
		"blee": issues.Issues{
			issues.New(issues.Root, config.OkLevel, "Blah"),
		},
	}

	ta.Rollup(o)
	b.AddSection("fred", o, ta)
	b.AddError(errors.New("boom"))
	s, err := b.ToYAML()

	assert.Nil(t, err)
	assert.Equal(t, reportYAML, s)
}

func TestBuilderJSON(t *testing.T) {
	b, ta := NewBuilder(internal.NewAliases()), NewTally()
	o := issues.Outcome{
		"blee": issues.Issues{
			issues.New(issues.Root, config.OkLevel, "Blah"),
		},
	}

	ta.Rollup(o)
	b.AddSection("fred", o, ta)
	b.AddError(errors.New("boom"))
	s, err := b.ToJSON()

	assert.Nil(t, err)
	assert.Equal(t, reportJSON, s)
}

func TestPrintSummary(t *testing.T) {
	b, ta := NewBuilder(internal.NewAliases()), NewTally()
	o := issues.Outcome{
		"blee": issues.Issues{
			issues.New(issues.Root, config.OkLevel, "Blah"),
		},
	}

	ta.Rollup(o)
	b.AddSection("fred", o, ta)
	b.AddError(errors.New("boom"))

	buff := bytes.NewBuffer([]byte(""))
	san := NewSanitizer(buff, false)
	b.PrintSummary(san)

	assert.Equal(t, summary, buff.String())
}

func TestPrintHeader(t *testing.T) {
	b, ta := NewBuilder(internal.NewAliases()), NewTally()
	o := issues.Outcome{
		"blee": issues.Issues{
			issues.New(issues.Root, config.OkLevel, "Blah"),
		},
	}

	ta.Rollup(o)
	b.AddSection("fred", o, ta)
	b.AddError(errors.New("boom"))

	buff := bytes.NewBuffer([]byte(""))
	san := NewSanitizer(buff, false)
	b.PrintHeader(san)

	assert.Equal(t, header, buff.String())
}

func TestPrintReport(t *testing.T) {
	b, ta := NewBuilder(internal.NewAliases()), NewTally()
	o := issues.Outcome{
		"blee": issues.Issues{
			issues.New(issues.Root, config.OkLevel, "Blah"),
		},
	}

	ta.Rollup(o)
	b.AddSection("fred", o, ta)
	b.AddError(errors.New("boom"))

	buff := bytes.NewBuffer([]byte(""))
	san := NewSanitizer(buff, false)
	b.PrintReport(config.OkLevel, san)

	assert.Equal(t, report, buff.String())
}

func TestTitleize(t *testing.T) {
	uu := map[string][]string{
		"cl":  {"CLUSTER", "CLUSTERS (1 SCANNED)"},
		"cm":  {"CONFIGMAP", "CONFIGMAPS (1 SCANNED)"},
		"dp":  {"DEPLOYMENT", "DEPLOYMENTS (1 SCANNED)"},
		"ds":  {"DAEMONSET", "DAEMONSETS (1 SCANNED)"},
		"hpa": {"HORIZONTALPODAUTOSCALER", "HORIZONTALPODAUTOSCALERS (1 SCANNED)"},
		"ing": {"INGRESS", "INGRESSES (1 SCANNED)"},
		"no":  {"NODE", "NODES (1 SCANNED)"},
		"np":  {"NETWORKPOLICY", "NETWORKPOLICIES (1 SCANNED)"},
		"ns":  {"NAMESPACE", "NAMESPACES (1 SCANNED)"},
		"pdb": {"PODDISRUPTIONBUDGET", "PODDISRUPTIONBUDGETS (1 SCANNED)"},
		"po":  {"POD", "PODS (1 SCANNED)"},
		"psp": {"PODSECURITYPOLICY", "PODSECURITYPOLICIES (1 SCANNED)"},
		"pv":  {"PERSISTENTVOLUME", "PERSISTENTVOLUMES (1 SCANNED)"},
		"pvc": {"PERSISTENTVOLUMECLAIM", "PERSISTENTVOLUMECLAIMS (1 SCANNED)"},
		"rs":  {"REPLICASET", "REPLICASETS (1 SCANNED)"},
		"sa":  {"SERVICEACCOUNT", "SERVICEACCOUNTS (1 SCANNED)"},
		"sec": {"SECRET", "SECRETS (1 SCANNED)"},
		"sts": {"STATEFULSET", "STATEFULSETS (1 SCANNED)"},
		"svc": {"SERVICE", "SERVICES (1 SCANNED)"},

		// Fallback
		"blee": {"BLEE", "BLEES (1 SCANNED)"},
	}

	a := internal.NewAliases()
	for k, e := range uu {
		assert.Equal(t, e[0], Titleize(a, k, 0))
		assert.Equal(t, e[1], Titleize(a, k, 1))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

var (
	reportJunit = "<testsuites name=\"Popeye\" tests=\"1\" failures=\"0\" errors=\"1\">\n\t<testsuite name=\"fred\" tests=\"1\" failures=\"0\" errors=\"0\">\n\t\t<properties>\n\t\t\t<property name=\"OK\" value=\"1\"></property>\n\t\t\t<property name=\"Info\" value=\"0\"></property>\n\t\t\t<property name=\"Warn\" value=\"0\"></property>\n\t\t\t<property name=\"Error\" value=\"0\"></property>\n\t\t\t<property name=\"Score\" value=\"100%\"></property>\n\t\t</properties>\n\t\t<testcase classname=\"\" name=\"blee\"></testcase>\n\t</testsuite>\n</testsuites>"

	reportJSON = "{\"popeye\":{\"score\":100,\"grade\":\"A\",\"sanitizers\":[{\"sanitizer\":\"fred\",\"tally\":{\"ok\":1,\"info\":0,\"warning\":0,\"error\":0,\"score\":100},\"issues\":{\"blee\":[{\"group\":\"__root__\",\"level\":0,\"message\":\"Blah\"}]}}],\"errors\":[{}]}}"

	reportYAML = `popeye:
  score: 100
  grade: A
  sanitizers:
  - sanitizer: fred
    tally:
      ok: 1
      info: 0
      warning: 0
      error: 0
      score: 100
    issues:
      blee:
      - group: __root__
        level: 0
        message: Blah
  errors:
  - {}
`

	summary = "\n\x1b[38;5;75;mSUMMARY\x1b[0m\n\x1b[38;5;75;m" + strings.Repeat("┅", 100) + "\x1b[0m\nYour cluster score: 100 -- A\n                                                                                \x1b[38;5;82;mo          .-'-.     \x1b[0m\n                                                                                \x1b[38;5;82;m o     __| A    `\\  \x1b[0m\n                                                                                \x1b[38;5;82;m  o   `-,-`--._   `\\\x1b[0m\n                                                                                \x1b[38;5;82;m []  .->'  a     `|-'\x1b[0m\n                                                                                \x1b[38;5;82;m  `=/ (__/_       /  \x1b[0m\n                                                                                \x1b[38;5;82;m    \\_,    `    _)  \x1b[0m\n                                                                                \x1b[38;5;82;m       `----;  |     \x1b[0m\n\n"
	header  = "\n\x1b[38;5;122;m ___     ___ _____   _____ \x1b[0m                                                     \x1b[38;5;75;mK          .-'-.     \x1b[0m\n\x1b[38;5;122;m| _ \\___| _ \\ __\\ \\ / / __|\x1b[0m                                                     \x1b[38;5;75;m 8     __|      `\\  \x1b[0m\n\x1b[38;5;122;m|  _/ _ \\  _/ _| \\ V /| _| \x1b[0m                                                     \x1b[38;5;75;m  s   `-,-`--._   `\\\x1b[0m\n\x1b[38;5;122;m|_| \\___/_| |___| |_| |___|\x1b[0m                                                     \x1b[38;5;75;m []  .->'  a     `|-'\x1b[0m\n\x1b[38;5;75;m  Biffs`em and Buffs`em!\x1b[0m                                                        \x1b[38;5;75;m  `=/ (__/_       /  \x1b[0m\n                                                                                \x1b[38;5;75;m    \\_,    `    _)  \x1b[0m\n                                                                                \x1b[38;5;75;m       `----;  |     \x1b[0m\n\n"
	report  = "\n\x1b[38;5;75;mFREDS (1 SCANNED)\x1b[0m" + strings.Repeat(" ", 60) + "💥 0 😱 0 🔊 0 ✅ 1 \x1b[38;5;122;m100\x1b[0m٪\n\x1b[38;5;75;m" + strings.Repeat("┅", 100) + "\x1b[0m\n  · \x1b[38;5;155;mblee\x1b[0m\x1b[38;5;250;m" + strings.Repeat(".", 91) + "\x1b[0m✅\n    ✅ \x1b[38;5;155;mBlah.\x1b[0m\n\n"
)
