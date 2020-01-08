package report

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
)

// TestSuites a collection of junit test suites.
type TestSuites struct {
	XMLName  xml.Name `xml:"testsuites"`
	Name     string   `xml:"name,attr"`
	Tests    int      `xml:"tests,attr"`
	Failures int      `xml:"failures,attr"`
	Errors   int      `xml:"errors,attr"`
	Suites   []TestSuite
}

// TestSuite represents a collection of tests
type TestSuite struct {
	XMLName    xml.Name   `xml:"testsuite"`
	Name       string     `xml:"name,attr"`
	Tests      int        `xml:"tests,attr"`
	Failures   int        `xml:"failures,attr"`
	Errors     int        `xml:"errors,attr"`
	Properties []Property `xml:"properties>property,omitempty"`
	TestCases  []TestCase
}

// TestCase represents a sing junit test.
type TestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	Classname string   `xml:"classname,attr"`
	Name      string   `xml:"name,attr"`
	Failures  []Failure
	Errors    []Error
}

// Property represents key/value pair.
type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Failure represents a test failure.
type Failure struct {
	XMLName xml.Name `xml:"failure"`
	Message string   `xml:"message,attr"`
	Type    string   `xml:"type,attr"`
}

// Error represents a test error..
type Error struct {
	XMLName xml.Name `xml:"error"`
	Message string   `xml:"message,attr"`
	Type    string   `xml:"type,attr"`
}

func junitMarshal(b *Builder) ([]byte, error) {
	s := TestSuites{
		Name:   "Popeye",
		Tests:  len(b.Report.Sections),
		Errors: len(b.Report.Errors),
	}

	for _, section := range b.Report.Sections {
		s.Suites = append(s.Suites, newSuite(section, b))
	}

	return xml.MarshalIndent(s, "", "\t")
}

func newSuite(s Section, b *Builder) TestSuite {
	total, fails, errs := numTests(s.Outcome)
	ts := TestSuite{
		Name:     b.aliases.FromAlias(s.Title),
		Tests:    total,
		Failures: fails,
		Errors:   errs,
	}
	ts.Properties = tallyToProps(s.Tally)

	for k, v := range s.Outcome {
		ts.TestCases = append(ts.TestCases, newTestCase(k, v))
	}
	return ts
}

func newTestCase(res string, ii issues.Issues) TestCase {
	ns, n := namespaced(res)
	tc := TestCase{
		Classname: ns,
		Name:      n,
	}

	for _, i := range ii {
		switch i.Level {
		case config.WarnLevel:
			tc.Failures = append(tc.Failures, newFailure(i))
		case config.ErrorLevel:
			tc.Errors = append(tc.Errors, newError(i))
		}
	}

	return tc
}

func numTests(o issues.Outcome) (total, fails, errors int) {
	for _, v := range o {
		total += len(v)
		for _, i := range v {
			if i.Level >= config.WarnLevel {
				fails++
			}
			if i.Level == config.ErrorLevel {
				errors++
			}
		}
	}
	return
}

func tallyToProps(t *Tally) []Property {
	var p []Property

	for i, s := range t.counts {
		p = append(p, newProp(indexToTally(i), strconv.Itoa(s)))
	}

	p = append(p, newProp("Score", fmt.Sprintf("%d%%", t.score)))

	return p
}

func namespaced(res string) (string, string) {
	tokens := strings.Split(res, "/")
	if len(tokens) < 2 {
		return "", res
	}
	return tokens[0], tokens[1]
}

func newFailure(i issues.Issue) Failure {
	return Failure{
		Message: i.Message,
		Type:    issues.LevelToStr(i.Level),
	}
}

func newError(i issues.Issue) Error {
	return Error{
		Message: i.Message,
		Type:    issues.LevelToStr(i.Level),
	}
}

func newProp(k, v string) Property {
	return Property{Name: k, Value: v}
}
