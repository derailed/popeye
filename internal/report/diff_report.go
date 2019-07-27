package report

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/derailed/popeye/internal/issues"
)

type deltaIssue struct {
	issues.Issue
	add bool
}

type deltaIssues []deltaIssue

func newDeltaIssue(i issues.Issue, add bool) deltaIssue {
	return deltaIssue{Issue: i, add: add}
}

type deltaOutcome map[string]deltaIssues

type resourceSection struct {
	outcomes deltaOutcome
	tallies  []DeltaScore
}

func newResourceSection() *resourceSection {
	return &resourceSection{
		outcomes: make(deltaOutcome),
	}
}

// DiffReport compares the last 2 sanitizer reports.
type DiffReport struct {
	r1, r2   *Report
	overall  DeltaScore
	sections map[string]*resourceSection
	errors   []error
}

func newDiffReport(r1, r2 *Report) *DiffReport {
	return &DiffReport{
		r1:       r1,
		r2:       r2,
		sections: make(map[string]*resourceSection),
	}
}

// Build computes diff between two sanitizer runs.
func (r *DiffReport) Build() {
	r.overall = NewDeltaScore(issues.OkLevel, r.r1.Score, r.r2.Score, false)
	for _, section := range r.r2.Sections {
		s := findSection(r.r1.Sections, section.Title)
		if s == nil {
			r.addError(fmt.Errorf("Unable to find matching sanitizer section `%s", section.Title))
		}
		r.addSection(*s, section)
	}
}

func findSection(ss []Section, title string) *Section {
	for _, s := range ss {
		if s.Title == title {
			return &s
		}
	}

	return nil
}

func (r *DiffReport) addError(err error) {
	r.errors = append(r.errors, err)
}

func (r *DiffReport) addSection(o, n Section) {
	s := &resourceSection{
		outcomes: make(deltaOutcome),
	}
	r.sections[n.Title] = s

	r.diffTallies(s, o.Tally, n.Tally)
	r.diffOutcome(n.Title, o.Outcome, n.Outcome)
}

// Namespaced return ns and name contained in given fqn.
func namespaced(fqn string) (string, string) {
	tokens := strings.Split(fqn, "/")
	if len(tokens) == 2 {
		return tokens[0], tokens[1]
	}
	return "", tokens[0]
}

func podName(n string) string {
	fmt.Println("Pod name", n)
	podNameRX := regexp.MustCompile(`\A.+/(\w+)-`)
	m := podNameRX.FindStringSubmatch(n)
	if len(m) < 2 {
		panic(fmt.Sprintf("Unable to deduce pod name from `%s", n))
	}
	return podNameRX.FindStringSubmatch(n)[1]
}

func (r *DiffReport) diffOutcome(section string, o, n issues.Outcome) {
	for k, newIssues := range n {
		oldIssues, ok := o[k]
		if !ok {
			// Pod may be reincarnated with a diff rs + id. Need to locate that pod.
			// Try a brute force pod name match for now. Could yield issues...
			po := podName(k)
			for k, ii := range o {
				if podName(k) == po {
					oldIssues = ii
					break
				}
			}
			if oldIssues == nil {
				r.addError(fmt.Errorf("Previous sanitizer missing resource ID `%s", k))
				continue
			}
		}
		fmt.Println("\n", k)
		r.sections[section].outcomes[k] = r.surface(oldIssues, newIssues)
	}
}

func (r *DiffReport) surface(o, n issues.Issues) deltaIssues {
	var ii deltaIssues
	for _, issue := range n {
		if issue.Level == issues.OkLevel {
			continue
		}
		if !inList(o, issue) {
			ii = append(ii, newDeltaIssue(issue, true))
		}
	}

	for _, issue := range o {
		if issue.Level == issues.OkLevel {
			continue
		}
		if !inList(n, issue) {
			ii = append(ii, newDeltaIssue(issue, false))
		}
	}

	return ii
}

func inList(ii issues.Issues, issue issues.Issue) bool {
	for _, i := range ii {
		if i == issue {
			return true
		}
	}
	return false
}

func (r *DiffReport) diffTallies(section *resourceSection, o, n *Tally) {
	for i, c := range n.counts {
		if c == o.counts[i] {
			continue
		}
		invert := true
		if i == len(n.counts)-1 {
			invert = false
		}
		section.tallies = append(section.tallies, NewDeltaScore(issues.Level(3-i), o.counts[i], c, invert))
	}
}
