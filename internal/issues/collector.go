package issues

// Collector represents a sanitizer issue container.
type Collector struct {
	outcomes Outcome
}

// NewCollector returns a new issue collector.
func NewCollector() *Collector {
	return &Collector{Outcome{}}
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

// AddSubOk add a sub ok issue.
func (c *Collector) AddSubOk(section, group, desc string) {
	c.addIssue(section, New(group, OkLevel, desc))
}

// AddSubOkf add a sub ok issue.
func (c *Collector) AddSubOkf(section, group, fmat string, args ...interface{}) {
	c.addIssue(section, Newf(group, OkLevel, fmat, args...))
}

// AddSubInfo add a sub info issue.
func (c *Collector) AddSubInfo(section, group, desc string) {
	c.addIssue(section, New(group, InfoLevel, desc))
}

// AddSubInfof add a sub info issue.
func (c *Collector) AddSubInfof(section, group, fmat string, args ...interface{}) {
	c.addIssue(section, Newf(group, InfoLevel, fmat, args...))
}

// AddSubWarn add a sub warning issue.
func (c *Collector) AddSubWarn(section, group, desc string) {
	c.addIssue(section, New(group, WarnLevel, desc))
}

// AddSubWarnf add a sub warning issue.
func (c *Collector) AddSubWarnf(section, group, fmat string, args ...interface{}) {
	c.addIssue(section, Newf(group, WarnLevel, fmat, args...))
}

// AddSubError add a sub error issue.
func (c *Collector) AddSubError(section, group, desc string) {
	c.addIssue(section, New(group, ErrorLevel, desc))
}

// AddSubErrorf add a sub error issue.
func (c *Collector) AddSubErrorf(section, group, fmat string, args ...interface{}) {
	c.addIssue(section, Newf(group, ErrorLevel, fmat, args...))
}

// AddErr adds a collection of errors.
func (c *Collector) AddErr(res string, errs ...error) {
	for _, e := range errs {
		c.addIssue(res, New(Root, ErrorLevel, e.Error()))
	}
}

// AddOk adds an ok issue.
func (c *Collector) AddOk(res, msg string) {
	c.addIssue(res, New(Root, OkLevel, msg))
}

// AddOkf adds an ok issue.
func (c *Collector) AddOkf(res, fmat string, args ...interface{}) {
	c.addIssue(res, Newf(Root, OkLevel, fmat, args...))
}

// AddInfo adds an info issue.
func (c *Collector) AddInfo(res, msg string) {
	c.addIssue(res, New(Root, InfoLevel, msg))
}

// AddInfof adds an info issue.
func (c *Collector) AddInfof(res, fmat string, args ...interface{}) {
	c.addIssue(res, Newf(Root, InfoLevel, fmat, args...))
}

// AddWarn adds a warning issue.
func (c *Collector) AddWarn(res, msg string) {
	c.addIssue(res, New(Root, WarnLevel, msg))
}

// AddWarnf adds a warning issue.
func (c *Collector) AddWarnf(res, fmat string, args ...interface{}) {
	c.addIssue(res, Newf(Root, WarnLevel, fmat, args...))
}

// AddError adds an error issue.
func (c *Collector) AddError(res, msg string) {
	c.addIssue(res, New(Root, ErrorLevel, msg))
}

// AddErrorf adds an info issue.
func (c *Collector) AddErrorf(res, fmat string, args ...interface{}) {
	c.addIssue(res, Newf(Root, ErrorLevel, fmat, args...))
}

// AddIssue adds 1 or more concerns to the collector.
func (c *Collector) addIssue(res string, concerns ...Issue) {
	if len(concerns) == 0 {
		return
	}
	c.outcomes[res] = append(c.outcomes[res], concerns...)
}
