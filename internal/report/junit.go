package report

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/popeye/internal/issues"
)

type TestSuites struct {
	XMLName  xml.Name `xml:"testsuites"`
	Name     string   `xml:"name,attr"`
	Tests    int      `xml:"tests,attr"`
	Failures int      `xml:"failures,attr"`
	Errors   int      `xml:"errors,attr"`
	Suites   []TestSuite
}

type TestSuite struct {
	XMLName    xml.Name   `xml:"testsuite"`
	Name       string     `xml:"name,attr"`
	Tests      int        `xml:"tests,attr"`
	Failures   int        `xml:"failures,attr"`
	Errors     int        `xml:"errors,attr"`
	Properties []Property `xml:"properties>property,omitempty"`
	TestCases  []TestCase
}

type TestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	Classname string   `xml:"classname,attr"`
	Name      string   `xml:"name,attr"`
	Failure   *Failure `xml:"failure,omitempty"`
}

type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type Failure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

func junitMarshal(b *Builder) ([]byte, error) {
	s := TestSuites{
		Name:   "Popeye",
		Tests:  len(b.Report.Sections),
		Errors: len(b.Report.Errors),
	}

	for _, section := range b.Report.Sections {
		s.Suites = append(s.Suites, newSuite(section))
	}

	return xml.MarshalIndent(s, "", "\t")
}

func newSuite(s Section) TestSuite {
	total, fails, errs := numTests(s.Outcome)
	ts := TestSuite{
		Name:     ResToTitle(s.Title),
		Tests:    total,
		Failures: fails,
		Errors:   errs,
	}
	ts.Properties = tallyToProps(s.Tally)

	for k, v := range s.Outcome {
		for _, issue := range v {
			ts.TestCases = append(ts.TestCases, newTestCase(k, issue))
		}

	}
	return ts
}

func numTests(o issues.Outcome) (total, fails, errors int) {
	for _, v := range o {
		total += len(v)
		for _, i := range v {
			if i.Level >= issues.WarnLevel {
				fails++
			}
			if i.Level == issues.ErrorLevel {
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

func newTestCase(res string, i issues.Issue) TestCase {
	ns, n := namespaced(res)
	return TestCase{
		Classname: ns,
		Name:      n,
		Failure:   newFailure(i),
	}
}

func namespaced(res string) (string, string) {
	tokens := strings.Split(res, "/")
	if len(tokens) < 2 {
		return "", res
	}
	return tokens[0], tokens[1]
}

func newFailure(i issues.Issue) *Failure {
	return &Failure{
		Message: i.Message,
		Type:    issues.LevelToStr(i.Level),
	}
}

func newProp(k, v string) Property {
	return Property{Name: k, Value: v}
}
