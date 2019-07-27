package issues

import "fmt"

// Collector represents a sanitizer issue container.
type Collector struct {
	outcomes Outcome
	codes    *Codes
}

// NewCollector returns a new issue collector.
func NewCollector(codes *Codes) *Collector {
	return &Collector{outcomes: Outcome{}, codes: codes}
}

// Outcome returns scan outcome.
func (c *Collector) Outcome() Outcome {
	return c.outcomes
}

// InitOutcome creates a places holder for potential issues.
func (c *Collector) InitOutcome(section string) {
	c.outcomes[section] = Issues{}
}

// NoConcerns returns true is scan is successful.
func (c *Collector) NoConcerns(section string) bool {
	return len(c.outcomes[section]) == 0
}

// MaxSeverity return the highest severity level foe the given section.
func (c *Collector) MaxSeverity(section string) Level {
	return c.outcomes.MaxSeverity(section)
}

// AddSubCode add a sub error code.
func (c *Collector) AddSubCode(code ID, section, group string, args ...interface{}) {
	co, ok := c.codes.Glossary[code]
	if !ok {
		panic(fmt.Sprintf("No code with ID %d", code))
	}
	c.addIssue(section, New(group, co.Severity, co.Format(code, args...)))
}

// AddCode add an error code.
func (c *Collector) AddCode(code ID, section string, args ...interface{}) {
	co, ok := c.codes.Glossary[code]
	if !ok {
		panic(fmt.Sprintf("No code with ID %d", code))
	}
	c.addIssue(section, New(Root, co.Severity, co.Format(code, args...)))
}

// AddErr adds a collection of errors.
func (c *Collector) AddErr(res string, errs ...error) {
	for _, e := range errs {
		c.addIssue(res, New(Root, ErrorLevel, e.Error()))
	}
}

// AddIssue adds 1 or more concerns to the collector.
func (c *Collector) addIssue(res string, concerns ...Issue) {
	if len(concerns) == 0 {
		return
	}
	c.outcomes[res] = append(c.outcomes[res], concerns...)
}
